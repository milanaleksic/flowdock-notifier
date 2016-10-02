package db

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// New creates new DB abstraction
func New() *DB {
	sess, err := session.NewSession()
	if err != nil {
		panic(err)
	}
	return &DB{
		dynamo: dynamodb.New(sess),
	}
}

// DB is an abstraction that keeps internals of working with the backend database
type DB struct {
	dynamo *dynamodb.DynamoDB
}

// IsActive returns true if the configuration table contains "active" configuration with value "true"
func (db *DB) IsActive() bool {
	resp, err := db.dynamo.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("flowdock-notifier"),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String("active"),
			},
		},
		AttributesToGet: []*string{
			aws.String("value"),
		},
	})
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if resp.Item != nil {
		return *resp.Item["value"].S == "true"
	}
	return false
}
