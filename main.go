package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func main() {
	fmt.Println("Hello World from Go!")
	fmt.Printf("Success, arguments received: %+v", os.Args)

	sess, err := session.NewSession()
	if err != nil {
		fmt.Println("failed to create session,", err)
		return
	}

	svc := dynamodb.New(sess)

	params := &dynamodb.DescribeTableInput{
		TableName: aws.String("flowdock-notifier"),
	}
	resp, err := svc.DescribeTable(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("Table description: %+v", resp)
}
