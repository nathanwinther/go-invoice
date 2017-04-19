package session

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/nathanwinther/go-uuid4"
	"invoice/config"
)

func Check(w http.ResponseWriter, r *http.Request) bool {
	uuid, ok := load(r)
	if !ok {
		return bounce(w, r)
	}

	db := dynamodb.New(config.DBSESS)

	resp, err := db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(config.DYNAMODB_PREFIX + "session"),
		Key: map[string]*dynamodb.AttributeValue{
			"uuid": {
				S: aws.String(uuid),
			},
		},
	})
	if err != nil {
		if !config.PRODUCTION {
			fmt.Println(err)
		}
	}

	if len(resp.Item) == 0 {
		writeCookie(w, "")
		return bounce(w, r)
	}

	modified, _ := strconv.ParseInt(*resp.Item["modified"].N, 10, 64)
	offset := time.Now().Unix() - modified

	if offset > int64(config.SESS_EXPIRES_REFRESH) {
		err = writeDb(uuid)
		if err != nil {
			if !config.PRODUCTION {
				fmt.Println(err)
			}
		}
	}

	writeCookie(w, uuid)

	return true
}

func New(w http.ResponseWriter) error {
	uuid, err := uuid4.New()
	if err != nil {
		return err
	}

	err = writeDb(uuid)
	if err != nil {
		return err
	}

	writeCookie(w, uuid)

	return nil
}

func bounce(w http.ResponseWriter, r *http.Request) bool {
	boomerang := base64.RawURLEncoding.EncodeToString([]byte(r.URL.Path))
	url := "/invoice/verify/" + boomerang

	http.Redirect(w, r, url, http.StatusFound)

	return false
}

func load(r *http.Request) (string, bool) {
	c, err := r.Cookie(config.SESS_NAME)
	if err != nil {
		return "", false
	}

	return c.Value, true
}

func writeCookie(w http.ResponseWriter, uuid string) {
	c := new(http.Cookie)
	c.Name = config.SESS_NAME
	c.Value = uuid
	c.Path = config.SESS_PATH
	if uuid != "" {
		c.MaxAge = config.SESS_EXPIRES
	} else {
		c.MaxAge = -1
	}
	c.Secure = config.ENV == "production"
	c.HttpOnly = true

	http.SetCookie(w, c)
}

func writeDb(uuid string) error {
	db := dynamodb.New(config.DBSESS)

	_, err := db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(config.DYNAMODB_PREFIX + "session"),
		Item: map[string]*dynamodb.AttributeValue{
			"uuid": {
				S: aws.String(uuid),
			},
			"modified": {
				N: aws.String(fmt.Sprintf("%d", time.Now().Unix())),
			},
		},
	})

	if err != nil {
		return err
	}

	return nil
}
