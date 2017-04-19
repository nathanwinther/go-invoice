package timesheet

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"invoice/config"
)

type Item struct {
	Enddate string
	Id      int
	Total   int
	Url     string
}

type Timesheet struct {
	Enddate   string
	Entries   []*Entry
	Id        int
	Selected  int
	Startdate string
	Total     int
}

type Entry struct {
	D        string
	DD       string
	DDD      string
	DDDD     string
	Key      string
	Hours    int
	M        string
	MM       string
	MMM      string
	MMMM     string
	Selected bool
	Today    bool
	YY       string
	YYYY     string
}

var (
	DATE_FORMAT = "2006-01-02"
)

func Items(company string) ([]*Item, error) {
	db := dynamodb.New(config.DBSESS)

	resp, err := db.Query(&dynamodb.QueryInput{
		TableName: aws.String(config.DYNAMODB_PREFIX + "timesheet"),
		AttributesToGet: []*string{
			aws.String("end_date"),
			aws.String("id"),
			aws.String("total"),
			aws.String("url"),
		},
		Limit: aws.Int64(10),
		KeyConditions: map[string]*dynamodb.Condition{
			"uuid": {
				ComparisonOperator: aws.String(dynamodb.ComparisonOperatorEq),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(company),
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	items := make([]*Item, *resp.Count)
	for idx, item := range resp.Items {
		items[idx] = new(Item)
		id, _ := strconv.Atoi(*item["id"].N)
		items[idx].Id = id
		items[idx].Enddate = *item["end_date"].S
		total, _ := strconv.Atoi(*item["total"].N)
		items[idx].Total = total
		items[idx].Url = *item["url"].S
	}

	return items, nil
}

func New(t1 *Timesheet) (*Timesheet, error) {
	t2 := new(Timesheet)

	t, err := time.Parse(DATE_FORMAT, t1.Enddate)
	if err != nil {
		return nil, err
	}

	t2.Entries = make([]*Entry, len(t1.Entries))
	for idx, _ := range t1.Entries {
		t = t.AddDate(0, 0, 1)

		e := new(Entry)
		e.Key = t.Format(DATE_FORMAT)
		e.Hours = 0
		e.YYYY = t.Format("2006")
		e.YY = t.Format("06")
		e.MMMM = t.Format("January")
		e.MMM = t.Format("Jan")
		e.MM = t.Format("01")
		e.M = t.Format("1")
		e.DDDD = t.Format("Monday")
		e.DDD = t.Format("Mon")
		e.DD = t.Format("02")
		e.D = t.Format("2")

		t2.Entries[idx] = e
	}

	t2.Id = t1.Id + 1
	t2.Startdate = t2.Entries[0].Key
	t2.Enddate = t2.Entries[len(t2.Entries)-1].Key

	return t2, nil
}

func (self *Timesheet) Index(key string) int {
	for idx, entry := range self.Entries {
		if entry.Key == key {
			return idx
		}
	}
	return 0
}

func (self *Timesheet) Save(company string) error {
	db := dynamodb.New(config.DBSESS)

	_, err := db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(config.DYNAMODB_PREFIX + "timesheet"),
		Item: map[string]*dynamodb.AttributeValue{
			"uuid": {
				S: aws.String(company),
			},
			"sort": {
				S: aws.String(self.SortKey()),
			},
			"end_date": {
				S: aws.String(self.Enddate),
			},
			"id": {
				N: aws.String(fmt.Sprintf("%d", self.Id)),
			},
			"modified": {
				S: aws.String(fmt.Sprintf("%d", time.Now().Unix())),
			},
			"start_date": {
				S: aws.String(self.Startdate),
			},
			"total": {
				N: aws.String(fmt.Sprintf("%d", self.Total)),
			},
			"url": {
				S: aws.String(fmt.Sprintf("/%s/%s.html", company, self.Enddate)),
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (self *Timesheet) SetHours(key string, hours string) {
	h, _ := strconv.Atoi(hours)
	self.Entries[self.Index(key)].Hours = h
	t := 0
	for idx, _ := range self.Entries {
		t += self.Entries[idx].Hours
	}
	self.Total = t
}

func (self *Timesheet) SetSelected(key string) {
	for idx, _ := range self.Entries {
		self.Entries[idx].Selected = false
	}

	if key == "" {
		key = time.Now().Format(DATE_FORMAT)
	}

	self.Selected = self.Index(key)
	self.Entries[self.Selected].Selected = true
}

func (self *Timesheet) SetToday() {
	key := time.Now().Format(DATE_FORMAT)
	for idx, entry := range self.Entries {
		self.Entries[idx].Today = false
		if entry.Key == key {
			self.Entries[idx].Today = true
		}
	}
}

func (self *Timesheet) SortKey() string {
	// Nines complement
	s := strings.Replace(self.Enddate, "-", "", -1)
	i, _ := strconv.Atoi(s)
	i = 99999999 - i
	return strconv.Itoa(i)
}
