package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/craigpastro/crudapp/instrumentation"
	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage"
	"github.com/kelseyhightower/envconfig"
)

const data = "some data"

var (
	ctx context.Context
	db  storage.Storage
)

type Config struct {
	DynamoDBRegion   string `envconfig:"DYNAMODB_REGION" default:"us-west-2"`
	DynamoDBEndpoint string `envconfig:"DYNAMODB_ENDPOINT" default:"http://localhost:8000"`
}

func TestMain(m *testing.M) {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		fmt.Println("error reading config", err)
		os.Exit(1)
	}

	err := createTable(config)
	if err != nil {
		fmt.Println("error creating DynamoDB table", err)
		os.Exit(1)
	}

	ctx = context.Background()
	tracer, _ := instrumentation.NewTracer(ctx, instrumentation.TracerConfig{Enabled: false})

	db, err = New(ctx, tracer, config.DynamoDBRegion, config.DynamoDBEndpoint)
	if err != nil {
		fmt.Println("error initializing DynamoDB", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func createTable(config Config) error {
	tableName := "Posts"

	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String(config.DynamoDBRegion),
		Endpoint: aws.String(config.DynamoDBEndpoint),
	})
	if err != nil {
		return fmt.Errorf("unable to initialize session: %w", err)
	}

	client := dynamodb.New(sess)

	input := &dynamodb.CreateTableInput{
		TableName:   aws.String(tableName),
		BillingMode: aws.String("PAY_PER_REQUEST"),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("UserID"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("PostID"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("UserID"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("PostID"),
				KeyType:       aws.String("RANGE"),
			},
		},
	}

	_, err = client.CreateTable(input)
	if err != nil {
		return fmt.Errorf("error creating table: %w", err)
	}

	return nil
}

func TestRead(t *testing.T) {
	userID := myid.New()
	postID, _, _ := db.Create(ctx, userID, data)
	record, err := db.Read(ctx, userID, postID)

	if err != nil {
		t.Errorf("error not nil: %v", err)
	}

	if record.UserID != userID {
		t.Errorf("wrong userID: got '%s', but wanted '%s'", record.UserID, userID)
	}

	if record.PostID != postID {
		t.Errorf("wrong postID: got '%s', but wanted '%s'", record.PostID, postID)
	}

	if record.Data != data {
		t.Errorf("wrong data: got '%s', but wanted '%s'", record.Data, data)
	}
}

func TestReadNotExists(t *testing.T) {
	userID := myid.New()
	_, err := db.Read(ctx, userID, "1")
	if err != storage.ErrPostDoesNotExist {
		t.Errorf("unexpected error: got '%v', but wanted '%v'", err, storage.ErrPostDoesNotExist)
	}
}

func TestReadAll(t *testing.T) {
	userID := myid.New()
	db.Create(ctx, userID, "data 1")
	db.Create(ctx, userID, "data 2")
	records, err := db.ReadAll(ctx, userID)

	if err != nil {
		t.Errorf("error not nil: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("wrong number of records: got '%d', but wanted '%d'", len(records), 2)
	}
}

func TestUpdate(t *testing.T) {
	userID := myid.New()
	postID, _, _ := db.Create(ctx, userID, data)
	newData := "new data"
	db.Update(ctx, userID, postID, newData)
	record, _ := db.Read(ctx, userID, postID)

	if record.Data != newData {
		t.Errorf("wrong data: got '%s', but wanted '%s'", record.Data, newData)
	}

	if record.CreatedAt.After(record.UpdatedAt) {
		t.Errorf("createdAt is after updatedAt")
	}
}

func TestUpdateNotExists(t *testing.T) {
	userID := myid.New()
	_, err := db.Update(ctx, userID, "1", "new data")
	if err != storage.ErrPostDoesNotExist {
		t.Errorf("unexpected error: got '%v', but wanted '%v'", err, storage.ErrPostDoesNotExist)
	}
}

func TestDelete(t *testing.T) {
	userID := myid.New()
	postID, _, _ := db.Create(ctx, userID, data)
	err := db.Delete(ctx, userID, postID)

	if err != nil {
		t.Errorf("error not nil: %v", err)
	}

	// Now try to read the deleted record; it should not exist.
	_, err = db.Read(ctx, userID, postID)
	if !errors.Is(err, storage.ErrPostDoesNotExist) {
		t.Errorf("unexpected error: got '%v', but wanted '%v'", err, storage.ErrPostDoesNotExist)
	}
}
