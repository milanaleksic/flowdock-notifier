package db

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const (
	messageSuffix = " Powered by [Igor](https://github.com/milanaleksic/igor)"
)

// New creates new DB abstraction
func New() *DB {
	sess, err := session.NewSession()
	if err != nil {
		panic(err)
	}
	db := &DB{
		dynamo: dynamodb.New(sess),
	}
	db.getActivity()
	return db
}

// DB is an abstraction that keeps internals of working with the backend database
type DB struct {
	dynamo                  *dynamodb.DynamoDB
	activeFrom, activeUntil time.Time
}

func (db *DB) getActivity() {
	resp, err := db.dynamo.Scan(&dynamodb.ScanInput{
		TableName:        aws.String("igor-config"),
		FilterExpression: aws.String("#id IN (:activeFrom, :activeUntil)"),
		ExpressionAttributeNames: map[string]*string{
			"#value": aws.String("value"),
			"#id":    aws.String("id"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":activeFrom": {
				S: aws.String("activeFrom"),
			},
			":activeUntil": {
				S: aws.String("activeUntil"),
			},
		},
		ProjectionExpression: aws.String("id, #value"),
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if len(resp.Items) != 2 {
		return
	}
	if resp.Items[0] != nil && resp.Items[1] != nil {
		var activeFrom, activeUntil string
		if *resp.Items[0]["id"].S == "activeFrom" {
			activeFrom = *resp.Items[0]["value"].S
			activeUntil = *resp.Items[1]["value"].S
		} else {
			activeFrom = *resp.Items[1]["value"].S
			activeUntil = *resp.Items[0]["value"].S
		}
		parsedActiveFrom, err := time.Parse(time.RFC822, activeFrom)
		if err != nil {
			log.Fatalf("Active from couldn't be parsed, err: %+v", err)
			return
		}
		parsedActiveUntil, err := time.Parse(time.RFC822, activeUntil)
		if err != nil {
			log.Fatalf("Active until couldn't be parsed, err: %+v", err)
			return
		}
		db.activeFrom = parsedActiveFrom
		db.activeUntil = parsedActiveUntil
	}
}

// IsActive returns true if the configuration table contains "active" configuration with value "true"
func (db *DB) IsActive() bool {
	return time.Now().Before(db.activeUntil) && time.Now().After(db.activeFrom)
}

// GetLastCommunicationWith returns when was the last time we talked to a user X
func (db *DB) GetLastCommunicationWith(username string) (*time.Time, error) {
	resp, err := db.dynamo.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("igor-communication"),
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
		TableName: aws.String("igor-communication"),
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

// GetActivationTimeStart returns the "activeFrom" from DynamoDB configuration
func (db *DB) GetActivationTimeStart() time.Time {
	return db.activeFrom
}

// GetResponseMessage will return the active reponse message
func (db *DB) GetResponseMessage() (string, error) {
	resp, err := db.dynamo.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("igor-config"),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String("message"),
			},
		},
		AttributesToGet: []*string{
			aws.String("value"),
		},
	})
	if err != nil {
		return "", err
	}
	if resp.Item == nil {
		log.Fatal("Seems that response message template is not available in the DB")
	}
	templ := *resp.Item["value"].S
	buff := new(bytes.Buffer)
	sweaters := struct {
		From  string
		Until string
	}{
		db.activeFrom.Format(time.RFC822),
		db.activeUntil.Format(time.RFC822),
	}
	tmpl, err := template.New("template").Parse(templ + messageSuffix)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(buff, sweaters)
	if err != nil {
		panic(err)
	}
	return buff.String(), nil
}
