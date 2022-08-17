package storage_test

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	ddb "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage"
	"github.com/craigpastro/crudapp/storage/dynamodb"
	"github.com/craigpastro/crudapp/storage/memory"
	"github.com/craigpastro/crudapp/storage/mongodb"
	"github.com/craigpastro/crudapp/storage/postgres"
	"github.com/craigpastro/crudapp/storage/redis"
	"github.com/craigpastro/crudapp/telemetry"
	realredis "github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

const data = "some data"

type storageTest struct {
	name     string
	storage  storage.Storage
	resource *dockertest.Resource
}

func TestStorage(t *testing.T) {
	dockerpool, err := dockertest.NewPool("")
	require.NoError(t, err)

	storageTests := []storageTest{
		newDynamoDB(t, dockerpool),
		newMemory(),
		newMongoDB(t, dockerpool),
		newPostgres(t, dockerpool),
		newRedis(t, dockerpool),
	}

	for _, test := range storageTests {
		t.Run(test.name, func(t *testing.T) {
			testRead(t, test.storage)
			testReadNotExists(t, test.storage)
			testReadAll(t, test.storage)
			testUpdate(t, test.storage)
			testUpdateNotExists(t, test.storage)
			testDelete(t, test.storage)
			testDeleteNotExists(t, test.storage)

			if test.resource != nil {
				if err := test.resource.Close(); err != nil {
					log.Println(err)
				}
			}
		})
	}
}

func newDynamoDB(t *testing.T, dockerpool *dockertest.Pool) storageTest {
	ctx := context.Background()
	tracer := telemetry.NewNoopTracer()

	resource, err := dockerpool.RunWithOptions(&dockertest.RunOptions{
		Repository: "amazon/dynamodb-local",
		Tag:        "latest",
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	require.NoError(t, err)

	var client *ddb.DynamoDB
	err = dockerpool.Retry(func() error {
		var err error
		client, err = dynamodb.CreateClient(ctx, dynamodb.Config{Region: "us-west-2", Port: resource.GetPort("8000/tcp")})
		if err != nil {
			return err
		}
		_, err = client.ListTables(&ddb.ListTablesInput{})
		return err
	})
	require.NoError(t, err)

	input := &ddb.CreateTableInput{
		TableName: aws.String(dynamodb.TableName),
		AttributeDefinitions: []*ddb.AttributeDefinition{
			{
				AttributeName: aws.String(dynamodb.UserIDAttribute),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String(dynamodb.PostIDAttribute),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*ddb.KeySchemaElement{
			{
				AttributeName: aws.String(dynamodb.UserIDAttribute),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String(dynamodb.PostIDAttribute),
				KeyType:       aws.String("RANGE"),
			},
		},
		BillingMode: aws.String("PAY_PER_REQUEST"),
	}
	if _, err := client.CreateTableWithContext(ctx, input); err != nil {
		if !strings.Contains(err.Error(), "Cannot create preexisting table") {
			log.Fatalf("error creating table: %v\n", err)
		}
	}

	return storageTest{
		name:     "dynamodb",
		storage:  dynamodb.New(client, tracer),
		resource: resource,
	}
}

func newMemory() storageTest {
	return storageTest{
		name:    "memory",
		storage: memory.New(telemetry.NewNoopTracer()),
	}
}

func newMongoDB(t *testing.T, dockerpool *dockertest.Pool) storageTest {
	ctx := context.Background()
	tracer := telemetry.NewNoopTracer()

	resource, err := dockerpool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mongo",
		Tag:        "latest",
		Env:        []string{"MONGO_INITDB_ROOT_USERNAME=mongodb", "MONGO_INITDB_ROOT_PASSWORD=password"},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	require.NoError(t, err)

	var coll *mongo.Collection
	err = dockerpool.Retry(func() error {
		var err error
		coll, err = mongodb.CreateCollection(ctx, mongodb.Config{URL: fmt.Sprintf("mongodb://mongodb:password@localhost:%s", resource.GetPort("27017/tcp"))})
		return err
	})
	require.NoError(t, err)

	return storageTest{
		name:     "mongodb",
		storage:  mongodb.New(coll, tracer),
		resource: resource,
	}
}

func newPostgres(t *testing.T, dockerpool *dockertest.Pool) storageTest {
	ctx := context.Background()
	tracer := telemetry.NewNoopTracer()

	resource, err := dockerpool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "latest",
		Env:        []string{"POSTGRES_USER=postgres", "POSTGRES_PASSWORD=password"},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	require.NoError(t, err)

	var pool *pgxpool.Pool
	err = dockerpool.Retry(func() error {
		var err error
		pool, err = postgres.CreatePool(ctx, postgres.Config{URL: fmt.Sprintf("postgres://postgres:password@localhost:%s/postgres", resource.GetPort("5432/tcp"))})
		if err != nil {
			return err
		}
		return pool.Ping(ctx)
	})
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS post (
		user_id TEXT NOT NULL,
		post_id TEXT NOT NULL,
		data TEXT,
		created_at TIMESTAMPTZ,
		updated_at TIMESTAMPTZ,
		PRIMARY KEY (user_id, post_id)
	)`)
	require.NoError(t, err)

	return storageTest{
		name:     "postgres",
		storage:  postgres.New(pool, tracer),
		resource: resource,
	}
}

func newRedis(t *testing.T, dockerpool *dockertest.Pool) storageTest {
	ctx := context.Background()
	tracer := telemetry.NewNoopTracer()

	resource, err := dockerpool.RunWithOptions(&dockertest.RunOptions{
		Repository: "redis",
		Tag:        "latest",
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	require.NoError(t, err)

	var client *realredis.Client
	err = dockerpool.Retry(func() error {
		var err error
		client, err = redis.CreateClient(ctx, redis.Config{Addr: fmt.Sprintf("localhost:%s", resource.GetPort("6379/tcp")), Password: ""})
		return err
	})
	require.NoError(t, err)

	return storageTest{
		name:     "redis",
		storage:  redis.New(client, tracer),
		resource: resource,
	}
}

func testRead(t *testing.T, storage storage.Storage) {
	ctx := context.Background()
	userID := myid.New()
	created, err := storage.Create(ctx, userID, data)
	require.NoError(t, err)
	record, err := storage.Read(ctx, created.UserID, created.PostID)
	require.NoError(t, err)

	require.Equal(t, record.UserID, created.UserID)
	require.Equal(t, record.PostID, created.PostID)
	require.Equal(t, record.Data, data)
}

func testReadNotExists(t *testing.T, db storage.Storage) {
	ctx := context.Background()
	userID := myid.New()

	_, err := db.Read(ctx, userID, "1")
	require.ErrorIs(t, err, storage.ErrPostDoesNotExist)
}

func testReadAll(t *testing.T, db storage.Storage) {
	ctx := context.Background()
	userID := myid.New()
	_, err := db.Create(ctx, userID, "data 1")
	require.NoError(t, err)
	_, err = db.Create(ctx, userID, "data 2")
	require.NoError(t, err)

	records, err := db.ReadAll(ctx, userID)
	require.NoError(t, err)

	require.Len(t, records, 2, "got '%d', want '%d'", len(records), 2)
}

func testUpdate(t *testing.T, db storage.Storage) {
	ctx := context.Background()
	userID := myid.New()
	created, err := db.Create(ctx, userID, data)
	require.NoError(t, err)

	newData := "new data"
	_, err = db.Update(ctx, userID, created.PostID, newData)
	require.NoError(t, err)
	record, err := db.Read(ctx, created.UserID, created.PostID)
	require.NoError(t, err)

	require.Equal(t, record.Data, newData, "got '%s', want '%s'")
	require.True(t, record.CreatedAt.Before(record.UpdatedAt))
}

func testUpdateNotExists(t *testing.T, db storage.Storage) {
	ctx := context.Background()
	userID := myid.New()

	_, err := db.Update(ctx, userID, "1", "new data")
	require.ErrorIs(t, err, storage.ErrPostDoesNotExist)
}

func testDelete(t *testing.T, db storage.Storage) {
	ctx := context.Background()
	userID := myid.New()
	created, _ := db.Create(ctx, userID, data)

	err := db.Delete(ctx, userID, created.PostID)
	require.NoError(t, err)

	// Now try to read the deleted record; it should not exist.
	_, err = db.Read(ctx, userID, created.PostID)
	require.ErrorIs(t, err, storage.ErrPostDoesNotExist)
}

func testDeleteNotExists(t *testing.T, db storage.Storage) {
	ctx := context.Background()
	userID := myid.New()
	postID := myid.New()

	err := db.Delete(ctx, userID, postID)
	require.NoError(t, err)
}
