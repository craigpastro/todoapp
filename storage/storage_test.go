package storage_test

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

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
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const data = "some data"

type storageTest struct {
	name      string
	storage   storage.Storage
	container testcontainers.Container
}

func TestStorage(t *testing.T) {
	storageTests := []storageTest{
		newDynamoDB(t),
		newMemory(),
		newMongoDB(t),
		newPostgres(t),
		newRedis(t),
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

			if test.container != nil {
				if err := test.container.Terminate(context.Background()); err != nil {
					log.Println(err)
				}
			}
		})
	}
}

func newDynamoDB(t *testing.T) storageTest {
	ctx := context.Background()
	tracer := telemetry.NewNoopTracer()

	req := testcontainers.ContainerRequest{
		Image:        "amazon/dynamodb-local:latest",
		ExposedPorts: []string{"8000/tcp"},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "8000/tcp")
	require.NoError(t, err)

	client, err := dynamodb.CreateClient(ctx, dynamodb.Config{Region: "us-west-2", Port: port.Port()})
	require.NoError(t, err)

	_, err = client.ListTables(&ddb.ListTablesInput{})
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
		name:      "dynamodb",
		storage:   dynamodb.New(client, tracer),
		container: container,
	}
}

func newMemory() storageTest {
	return storageTest{
		name:    "memory",
		storage: memory.New(telemetry.NewNoopTracer()),
	}
}

func newMongoDB(t *testing.T) storageTest {
	ctx := context.Background()
	tracer := telemetry.NewNoopTracer()

	req := testcontainers.ContainerRequest{
		Image:        "mongo:latest",
		ExposedPorts: []string{"27017/tcp"},
		Env:          map[string]string{"MONGO_INITDB_ROOT_USERNAME": "mongodb", "MONGO_INITDB_ROOT_PASSWORD": "password"},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "27017/tcp")
	require.NoError(t, err)

	coll, err := mongodb.CreateCollection(ctx, mongodb.Config{URL: fmt.Sprintf("mongodb://mongodb:password@localhost:%s", port.Port())})
	require.NoError(t, err)

	return storageTest{
		name:      "mongodb",
		storage:   mongodb.New(coll, tracer),
		container: container,
	}
}

func newPostgres(t *testing.T) storageTest {
	ctx := context.Background()
	tracer := telemetry.NewNoopTracer()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env:          map[string]string{"POSTGRES_USER": "postgres", "POSTGRES_PASSWORD": "password"},
		WaitingFor:   wait.ForLog("database system is ready to accept connections"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)

	pool, err := postgres.CreatePool(ctx, postgres.Config{URL: fmt.Sprintf("postgres://postgres:password@localhost:%s/postgres", port.Port())})
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
		name:      "postgres",
		storage:   postgres.New(pool, tracer),
		container: container,
	}
}

func newRedis(t *testing.T) storageTest {
	ctx := context.Background()
	tracer := telemetry.NewNoopTracer()

	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "6379/tcp")
	require.NoError(t, err)

	client, err := redis.CreateClient(ctx, redis.Config{Addr: fmt.Sprintf("localhost:%s", port.Port()), Password: ""})
	require.NoError(t, err)

	return storageTest{
		name:      "redis",
		storage:   redis.New(client, tracer),
		container: container,
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

	time.Sleep(time.Millisecond) // just in case
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
