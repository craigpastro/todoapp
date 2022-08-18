package dynamodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/cenkalti/backoff/v4"
	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage"
	"go.opentelemetry.io/otel/trace"
)

const (
	TableName       = "Posts"
	UserIDAttribute = "UserID"
	PostIDAttribute = "PostID"
)

type DynamoDB struct {
	client *dynamodb.DynamoDB
	tracer trace.Tracer
}

type Config struct {
	Region string
	// Port. If is supplied it means that we are running locally.
	Port string
}

func New(client *dynamodb.DynamoDB, tracer trace.Tracer) *DynamoDB {
	return &DynamoDB{
		client: client,
		tracer: tracer,
	}
}

func CreateClient(ctx context.Context, config Config) (*dynamodb.DynamoDB, error) {
	cfg := aws.Config{Region: aws.String(config.Region)}
	if config.Port != "" {
		cfg.Endpoint = aws.String(fmt.Sprintf("http://localhost:%s", config.Port))
	}
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            cfg,
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, fmt.Errorf("error initializing DynamoDB: %w", err)
	}

	client := dynamodb.New(sess)
	backoff.Retry(func() error {
		_, err := client.ListTables(&dynamodb.ListTablesInput{})
		return err
	}, backoff.NewExponentialBackOff())

	return client, nil
}

func (d *DynamoDB) Create(ctx context.Context, userID, data string) (*storage.Record, error) {
	ctx, span := d.tracer.Start(ctx, "dynamodb.Create")
	defer span.End()

	postID := myid.New()
	now := time.Now()
	record := storage.NewRecord(userID, postID, data, now, now)
	av, err := dynamodbattribute.MarshalMap(record)
	if err != nil {
		return nil, fmt.Errorf("error marshalling: %w", err)
	}

	if _, err := d.client.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(TableName),
	}); err != nil {
		return nil, fmt.Errorf("error creating: %w", err)
	}

	return record, nil
}

func (d *DynamoDB) Read(ctx context.Context, userID, postID string) (*storage.Record, error) {
	ctx, span := d.tracer.Start(ctx, "dynamodb.Read")
	defer span.End()

	result, err := d.client.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(TableName),
		Key:       createKey(userID, postID),
	})
	if err != nil {
		return nil, fmt.Errorf("error reading: %w", err)
	}
	if result.Item == nil {
		return nil, storage.ErrPostDoesNotExist
	}

	var record storage.Record
	if err := dynamodbattribute.UnmarshalMap(result.Item, &record); err != nil {
		return nil, fmt.Errorf("error unmarshaling, %w", err)
	}

	return &record, nil
}

func (d *DynamoDB) ReadAll(ctx context.Context, userID string) ([]*storage.Record, error) {
	ctx, span := d.tracer.Start(ctx, "dynamodb.ReadAll")
	defer span.End()

	keyCond := expression.Key(UserIDAttribute).Equal(expression.Value(userID))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return nil, fmt.Errorf("error building expression: %w", err)
	}
	result, err := d.client.QueryWithContext(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(TableName),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		return nil, fmt.Errorf("error reading all: %w", err)
	}

	var res []*storage.Record
	if err := dynamodbattribute.UnmarshalListOfMaps(result.Items, &res); err != nil {
		return nil, fmt.Errorf("error unmarshaling, %w", err)
	}

	return res, nil
}

func (d *DynamoDB) Update(ctx context.Context, userID, postID, data string) (time.Time, error) {
	ctx, span := d.tracer.Start(ctx, "dynamodb.Update")
	defer span.End()

	now := time.Now()
	update := expression.
		Set(expression.Name("Data"), expression.Value(data)).
		Set(expression.Name("UpdatedAt"), expression.Value(now))
	condition := expression.AttributeExists(expression.Name(UserIDAttribute)).And(expression.AttributeExists(expression.Name(PostIDAttribute)))
	expr, err := expression.NewBuilder().WithUpdate(update).WithCondition(condition).Build()
	if err != nil {
		return time.Time{}, fmt.Errorf("error building expression: %w", err)
	}

	input := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(TableName),
		Key:                       createKey(userID, postID),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ConditionExpression:       expr.Condition(),
	}
	if _, err := d.client.UpdateItemWithContext(ctx, input); err != nil {
		if t, ok := err.(awserr.Error); ok && t.Code() == "ConditionalCheckFailedException" {
			return time.Time{}, storage.ErrPostDoesNotExist
		}

		return time.Time{}, fmt.Errorf("error creating: %w", err)
	}

	return now, nil
}

func (d *DynamoDB) Delete(ctx context.Context, userID, postID string) error {
	ctx, span := d.tracer.Start(ctx, "dynamodb.Delete")
	defer span.End()

	if _, err := d.client.DeleteItemWithContext(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(TableName),
		Key:       createKey(userID, postID),
	}); err != nil {
		return fmt.Errorf("error deleting, %w", err)
	}

	return nil
}

func createKey(userID, postID string) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{
		UserIDAttribute: {
			S: aws.String(userID),
		},
		PostIDAttribute: {
			S: aws.String(postID),
		},
	}
}
