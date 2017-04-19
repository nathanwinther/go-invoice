package helper

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func String(item map[string]*dynamodb.AttributeValue, key string) string {
	value, ok := item[key]
	if !ok {
		return ""
	}
	return *value.S
}

func Number(item map[string]*dynamodb.AttributeValue, key string) string {
	value, ok := item[key]
	if !ok {
		return "0"
	}
	return *value.N
}
