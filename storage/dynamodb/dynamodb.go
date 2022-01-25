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
	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage"
	"go.opentelemetry.io/otel/trace"
)

const tableName = "Posts"

type DynamoDB struct {
	client *dynamodb.DynamoDB
	tracer trace.Tracer
}

func New(ctx context.Context, tracer trace.Tracer, region string, local bool) (storage.Storage, error) {
	config := aws.Config{Region: aws.String(region)}
	if local {
		config.Endpoint = aws.String("http://localhost:8000")
	}
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            config,
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to initialize session: %w", err)
	}

	return &DynamoDB{
		client: dynamodb.New(sess),
		tracer: tracer,
	}, nil
}

func (d *DynamoDB) Create(ctx context.Context, userID, data string) (string, time.Time, error) {
	ctx, span := d.tracer.Start(ctx, "dynamodb.Create")
	defer span.End()

	postID := myid.New()
	now := time.Now()
	record := storage.NewRecord(userID, postID, data, now, now)
	av, err := dynamodbattribute.MarshalMap(record)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("error marshalling: %v", err)
	}

	if _, err = d.client.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}); err != nil {
		return "", time.Time{}, fmt.Errorf("error creating: %v", err)
	}

	return postID, now, nil
}

func (d *DynamoDB) Read(ctx context.Context, userID, postID string) (*storage.Record, error) {
	ctx, span := d.tracer.Start(ctx, "dynamodb.Read")
	defer span.End()

	result, err := d.client.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key:       createKey(userID, postID),
	})
	if err != nil {
		return nil, fmt.Errorf("error reading: %v", err)
	}
	if result.Item == nil {
		return nil, storage.ErrPostDoesNotExist
	}

	var record storage.Record
	if err = dynamodbattribute.UnmarshalMap(result.Item, &record); err != nil {
		return nil, fmt.Errorf("error unmarshaling, %v", err)
	}

	return &record, nil
}

func (d *DynamoDB) ReadAll(ctx context.Context, userID string) ([]*storage.Record, error) {
	ctx, span := d.tracer.Start(ctx, "dynamodb.ReadAll")
	defer span.End()

	keyCond := expression.Key("UserID").Equal(expression.Value(userID))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return nil, fmt.Errorf("error building expression: %v", err)
	}
	rows, err := d.client.QueryWithContext(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(tableName),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		return nil, fmt.Errorf("error reading all: %v", err)
	}

	res := []*storage.Record{}
	for _, row := range rows.Items {
		var record storage.Record
		if err = dynamodbattribute.UnmarshalMap(row, &record); err != nil {
			return nil, fmt.Errorf("error unmarshaling, %v", err)
		}
		res = append(res, &record)
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
	condition := expression.AttributeExists(expression.Name("UserID")).And(expression.AttributeExists(expression.Name("PostID")))
	expr, err := expression.NewBuilder().WithUpdate(update).WithCondition(condition).Build()
	if err != nil {
		return time.Time{}, fmt.Errorf("error building expression: %v", err)
	}

	input := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(tableName),
		Key:                       createKey(userID, postID),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ConditionExpression:       expr.Condition(),
	}
	if _, err = d.client.UpdateItemWithContext(ctx, input); err != nil {
		if t, ok := err.(awserr.Error); ok && t.Code() == "ConditionalCheckFailedException" {
			return time.Time{}, storage.ErrPostDoesNotExist
		}

		return time.Time{}, fmt.Errorf("error creating: %v", err)
	}

	return now, nil
}

func (d *DynamoDB) Delete(ctx context.Context, userID, postID string) error {
	ctx, span := d.tracer.Start(ctx, "dynamodb.Delete")
	defer span.End()

	if _, err := d.client.DeleteItemWithContext(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key:       createKey(userID, postID),
	}); err != nil {
		return fmt.Errorf("error deleting, %v", err)
	}

	return nil
}

func createKey(userID, postID string) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{
		"UserID": {
			S: aws.String(userID),
		},
		"PostID": {
			S: aws.String(postID),
		},
	}
}
