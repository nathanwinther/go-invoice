package company

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"strconv"
	"time"
)

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"invoice/config"
	"invoice/helper"
	"invoice/timesheet"
)

type Item struct {
	Name string
	Uuid string
}

type Company struct {
	Modified  int64
	Name      string
	Rate      float64
	Template  string
	Text1     string
	Text2     string
	Text3     string
	Text4     string
	Timesheet *timesheet.Timesheet
	Uuid      string
}

func Items() ([]*Item, error) {
	db := dynamodb.New(config.DBSESS)

	resp, err := db.Scan(&dynamodb.ScanInput{
		TableName: aws.String(config.DYNAMODB_PREFIX + "company"),
		AttributesToGet: []*string{
			aws.String("company"),
			aws.String("uuid"),
		},
	})
	if err != nil {
		return nil, err
	}

	items := make([]*Item, *resp.Count)
	for idx, item := range resp.Items {
		items[idx] = new(Item)
		items[idx].Name = helper.String(item, "company")
		items[idx].Uuid = helper.String(item, "uuid")
	}

	return items, nil
}

func Load(uuid string, key string) (*Company, error) {
	db := dynamodb.New(config.DBSESS)

	resp, err := db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(config.DYNAMODB_PREFIX + "company"),
		Key: map[string]*dynamodb.AttributeValue{
			"uuid": {
				S: aws.String(uuid),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	c := new(Company)

	modified, _ := strconv.ParseInt(helper.Number(resp.Item, "modified"), 10, 64)
	c.Modified = modified
	c.Name = helper.String(resp.Item, "company")
	rate, _ := strconv.ParseFloat(helper.String(resp.Item, "rate"), 64)
	c.Rate = rate
	c.Template = helper.String(resp.Item, "template")
	c.Text1 = helper.String(resp.Item, "text1")
	c.Text2 = helper.String(resp.Item, "text2")
	c.Text3 = helper.String(resp.Item, "text3")
	c.Text4 = helper.String(resp.Item, "text4")
	t := new(timesheet.Timesheet)
	err = json.Unmarshal([]byte(helper.String(resp.Item, "timesheet")), t)
	if err != nil {
		return nil, err
	}
	c.Timesheet = t
	c.Timesheet.SetToday()
	c.Timesheet.SetSelected(key)
	c.Uuid = helper.String(resp.Item, "uuid")

	return c, nil
}

func (self *Company) CloseTimesheet() error {
	// Render HTML
	html, err := self.Html()
	if err != nil {
		return err
	}

	// Save copy to S3
	sess, err := config.NewS3Session()
	if err != nil {
		return err
	}

	svc := s3.New(sess)

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(config.S3_BUCKET),
		Key:         aws.String(fmt.Sprintf("%s/%s.html", self.Uuid, self.Timesheet.Enddate)),
		Body:        bytes.NewReader(html.Bytes()),
		ContentType: aws.String("text/html"),
	})
	if err != nil {
		return err
	}

	// Save to Dynamodb
	err = self.Timesheet.Save(self.Uuid)
	if err != nil {
		return err
	}

	return nil
}

func (self *Company) Html() (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	funcs := template.FuncMap{
		"money": func(rate float64, hours int) string {
			total := rate * float64(hours)
			return fmt.Sprintf("$%.2f", total)
		},
	}

	tpl, err := template.New("__RUNTIME__").Funcs(funcs).Parse(self.Template)
	if err != nil {
		return buf, err
	}

	m := map[string]interface{}{
		"Company": self,
	}

	err = tpl.Execute(buf, m)
	if err != nil {
		return buf, err
	}

	return buf, nil
}

func (self *Company) Save() error {
	timesheet, err := json.Marshal(self.Timesheet)
	if err != nil {
		return err
	}

	db := dynamodb.New(config.DBSESS)

	items := map[string]*dynamodb.AttributeValue{
		"modified": {
			N: aws.String(fmt.Sprintf("%d", time.Now().Unix())),
		},
		"rate": {
			S: aws.String(fmt.Sprintf("%f", self.Rate)),
		},
		"timesheet": {
			S: aws.String(string(timesheet)),
		},
		"uuid": {
			S: aws.String(self.Uuid),
		},
	}

	if self.Name != "" {
		items["company"] = &dynamodb.AttributeValue{
			S: aws.String(self.Name),
		}
	}

	if self.Template != "" {
		items["template"] = &dynamodb.AttributeValue{
			S: aws.String(self.Template),
		}
	}

	if self.Text1 != "" {
		items["text1"] = &dynamodb.AttributeValue{
			S: aws.String(self.Text1),
		}
	}

	if self.Text2 != "" {
		items["text2"] = &dynamodb.AttributeValue{
			S: aws.String(self.Text2),
		}
	}

	if self.Text3 != "" {
		items["text3"] = &dynamodb.AttributeValue{
			S: aws.String(self.Text3),
		}
	}

	if self.Text4 != "" {
		items["text4"] = &dynamodb.AttributeValue{
			S: aws.String(self.Text4),
		}
	}

	_, err = db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(config.DYNAMODB_PREFIX + "company"),
		Item:      items,
	})
	if err != nil {
		return err
	}

	return nil
}
