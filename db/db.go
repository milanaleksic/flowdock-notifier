package db

import (
	"fmt"
	"time"

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

// GetLastCommunicationWith returns when was the last time we talked to a user X
func (db *DB) GetLastCommunicationWith(username string) (*time.Time, error) {
	resp, err := db.dynamo.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("flowdock-notifier-communication"),
		Key: map[string]*dynamodb.AttributeValue{
			"userid": {
				S: aws.String(username),
			},
		},
		AttributesToGet: []*string{
			aws.String("moment"),
		},
	})
	if err != nil {
		return nil, err
	}
	if resp.Item == nil {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, *resp.Item["moment"].S)
	return &parsed, err
}

// SetLastCommunicationWith sets the last time we communicated with some Flowdock user
func (db *DB) SetLastCommunicationWith(username string, moment time.Time) error {
	_, err := db.dynamo.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("flowdock-notifier-communication"),
		Item: map[string]*dynamodb.AttributeValue{
			"userid": {
				S: aws.String(username),
			},
			"moment": {
				S: aws.String(moment.Format(time.RFC3339)),
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}
