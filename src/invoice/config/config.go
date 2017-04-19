package config

/*
 * Expected ENVIRONMENT variables
 *
 * APP_ENV
 * APP_DYNAMODB_PREFIX
 * APP_DYNAMODB_PROFILE
 * APP_DYNAMODB_REGION
 * APP_DYNAMODB_VERSION
 * APP_S3_BUCKET
 * APP_S3_PROFILE
 * APP_S3_REGION
 * APP_S3_VERSION
 * APP_S3_WEBSITE
 * APP_SESS_NAME
 * APP_SESS_EXPIRES
 * APP_SESS_EXPIRES_REFRESH
 * APP_SESS_PATH
 * APP_TOTP_SECRET
 */

import (
	"os"
	"strconv"
)

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

var (
	ENV                  = os.Getenv("APP_ENV")
	DYNAMODB_PREFIX      = os.Getenv("APP_DYNAMODB_PREFIX")
	DYNAMODB_PROFILE     = os.Getenv("APP_DYNAMODB_PROFILE")
	DYNAMODB_REGION      = os.Getenv("APP_DYNAMODB_REGION")
	DYNAMODB_VERSION     = os.Getenv("APP_DYNAMODB_VERSION")
	PRODUCTION           = false
	S3_BUCKET            = os.Getenv("APP_S3_BUCKET")
	S3_PROFILE           = os.Getenv("APP_S3_PROFILE")
	S3_REGION            = os.Getenv("APP_S3_REGION")
	S3_VERSION           = os.Getenv("APP_S3_VERSION")
	S3_WEBSITE           = os.Getenv("APP_S3_WEBSITE")
	SESS_NAME            = os.Getenv("APP_SESS_NAME")
	SESS_EXPIRES         = 0
	SESS_EXPIRES_REFRESH = 0
	SESS_PATH            = os.Getenv("APP_SESS_PATH")
	TOTP_SECRET          = os.Getenv("APP_TOTP_SECRET")
)

var (
	DBSESS *session.Session
)

func init() {
	if ENV == "production" {
		PRODUCTION = true
	}

	SESS_EXPIRES, _ = strconv.Atoi(os.Getenv("APP_SESS_EXPIRES"))
	SESS_EXPIRES_REFRESH, _ = strconv.Atoi(os.Getenv("APP_SESS_EXPIRES_REFRESH"))

	// Init AWS session for DynamoDB
	s, err := NewDbSession()
	if err != nil {
		panic(err)
	}

	DBSESS = s
}

func NewDbSession() (*session.Session, error) {
	shared := os.Getenv("HOME") + "/.aws/credentials"

	cred := credentials.NewSharedCredentials(shared, DYNAMODB_PROFILE)

	sess, err := session.NewSession(&aws.Config{
		Credentials: cred,
		Region:      aws.String(DYNAMODB_REGION),
	})
	if err != nil {
		return nil, err
	}

	return sess, nil
}

func NewS3Session() (*session.Session, error) {
	shared := os.Getenv("HOME") + "/.aws/credentials"

	cred := credentials.NewSharedCredentials(shared, S3_PROFILE)

	sess, err := session.NewSession(&aws.Config{
		Credentials: cred,
		Region:      aws.String(S3_REGION),
	})
	if err != nil {
		return nil, err
	}

	return sess, nil
}
