package table

import (
	"fmt"
)

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"invoice/config"
)

func Delete() {
	fmt.Println(config.ENV)
	deleteTables()
}

func Create() {
	fmt.Println(config.ENV)
	createTables()
}

func Describe() {
	fmt.Println(config.ENV)
	describeTables()
}

func describeTables() {
	db := dynamodb.New(config.DBSESS)

	tables := []string{
		"company",
		"session",
		"timesheet",
	}

	for _, table := range tables {
		resp, err := db.DescribeTable(&dynamodb.DescribeTableInput{
			TableName: aws.String(config.DYNAMODB_PREFIX + table),
		})

		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(resp)
	}

}

func deleteTables() {
	db := dynamodb.New(config.DBSESS)

	tables := []string{
		"company",
		"session",
		"timesheet",
	}

	for _, table := range tables {
		resp, err := db.DeleteTable(&dynamodb.DeleteTableInput{
			TableName: aws.String(config.DYNAMODB_PREFIX + table),
		})

		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(resp)
	}
}

func createTables() {
	db := dynamodb.New(config.DBSESS)

	// Company
	resp, err := db.CreateTable(&dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			&dynamodb.AttributeDefinition{
				AttributeName: aws.String("uuid"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			&dynamodb.KeySchemaElement{
				AttributeName: aws.String("uuid"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
		TableName: aws.String(config.DYNAMODB_PREFIX + "company"),
	})

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(resp)

	// Session
	resp, err = db.CreateTable(&dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			&dynamodb.AttributeDefinition{
				AttributeName: aws.String("uuid"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			&dynamodb.KeySchemaElement{
				AttributeName: aws.String("uuid"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
		TableName: aws.String(config.DYNAMODB_PREFIX + "session"),
	})

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(resp)

	// Invoice
	resp, err = db.CreateTable(&dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			&dynamodb.AttributeDefinition{
				AttributeName: aws.String("uuid"),
				AttributeType: aws.String("S"),
			},
			&dynamodb.AttributeDefinition{
				AttributeName: aws.String("sort"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			&dynamodb.KeySchemaElement{
				AttributeName: aws.String("uuid"),
				KeyType:       aws.String("HASH"),
			},
			&dynamodb.KeySchemaElement{
				AttributeName: aws.String("sort"),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
		TableName: aws.String(config.DYNAMODB_PREFIX + "timesheet"),
	})

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(resp)
}
