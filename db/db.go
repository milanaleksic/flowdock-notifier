package db

import (
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/milanaleksic/igor"
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

// SetLastCommunicationWith sets the last time we communicated with some Flowdock user
func (db *DB) SetLastCommunicationWith(userConfig *igor.UserConfig, username string, moment time.Time) error {
	_, err := db.dynamo.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("igor"),
		Key: map[string]*dynamodb.AttributeValue{
			"userId": {S: aws.String(userConfig.Identity)},
		},
		UpdateExpression:          aws.String("SET lastCommunication.#username = :new"),
		ExpressionAttributeNames:  map[string]*string{"#username": aws.String(username)},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{":new": {S: aws.String(moment.Format(time.RFC3339))}},
	})
	if err != nil {
		_, err = db.dynamo.UpdateItem(&dynamodb.UpdateItemInput{
			TableName: aws.String("igor"),
			Key: map[string]*dynamodb.AttributeValue{
				"userId": {S: aws.String(userConfig.Identity)},
			},
			UpdateExpression: aws.String("SET lastCommunication = if_not_exists(lastCommunication, :new)"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{":new": {M: map[string]*dynamodb.AttributeValue{
				username: &dynamodb.AttributeValue{S: aws.String(moment.Format(time.RFC3339))},
			}}},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) GetAllConfigs() (allConfigs []*igor.UserConfig, err error) {
	allConfigs = make([]*igor.UserConfig, 0)
	resp, err := db.dynamo.Scan(&dynamodb.ScanInput{
		TableName:      aws.String("igor"),
		ConsistentRead: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}
	if resp.Items == nil {
		return nil, nil
	}
	for _, item := range resp.Items {
		identity := *item["userId"].S
		messageFormat := *item["message"].S
		flowdockUsername := *item["flowdockUsername"].S
		flowdockToken := *item["flowdockToken"].S
		lastCommunication := make(map[string]time.Time)
		if commMap, ok := item["lastCommunication"]; ok {
			lastCommunicationMap := commMap.M
			for user, lastTime := range lastCommunicationMap {
				lastTimeParsed, err := time.Parse(time.RFC3339, *lastTime.S)
				if err != nil {
					log.Fatalf("Last time %s couldn't be parsed, err: %+v", lastTime, err)
					return nil, err
				}
				lastCommunication[user] = lastTimeParsed
			}
		}
		parsedActiveFrom, err := time.Parse(time.RFC3339, *item["activeFrom"].S)
		if err != nil {
			log.Fatalf("Active from couldn't be parsed, err: %+v", err)
			return nil, err
		}
		parsedActiveUntil, err := time.Parse(time.RFC3339, *item["activeUntil"].S)
		if err != nil {
			log.Fatalf("Active until couldn't be parsed, err: %+v", err)
			return nil, err
		}
		newConfig := igor.New(identity, messageFormat, flowdockUsername, flowdockToken, parsedActiveFrom, parsedActiveUntil, lastCommunication)
		allConfigs = append(allConfigs, newConfig)
	}
	return allConfigs, err
}
