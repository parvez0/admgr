package tests_test

import (
	"fmt"
	"github.com/Pallinder/go-randomdata"
	"github.com/bluele/factory-go/factory"
	"github.com/google/uuid"
	"github.com/kiran-anand14/admgr/internal/pkg/api"
	"github.com/kiran-anand14/admgr/internal/pkg/models"
	"github.com/kiran-anand14/admgr/internal/pkg/storage/mysql"
	"github.com/stretchr/testify/assert"
	"strconv"
	"strings"
	"testing"
	"time"
)

type SlotFactory struct {
	Status     []string
	Instances  int
	DateFormat string
}

func (s *SlotFactory) WithStatus(status []string) *SlotFactory {
	s.Status = status
	return s
}

func (s *SlotFactory) WithInstances(n int) *SlotFactory {
	s.Instances = n
	return s
}

func (s *SlotFactory) WithDateFormat(format string) *SlotFactory {
	s.DateFormat = format
	return s
}

func (s *SlotFactory) fill_default() *SlotFactory {
	if s.Instances == 0 {
		s.Instances = 1
	}
	if s.Status == nil {
		s.Status = []string{
			models.SlotStatusOpen,
			models.SlotStatusBooked,
			models.SlotStatusHold,
			models.SlotStatusClosed,
		}
	}
	if s.DateFormat == "" {
		s.DateFormat = "2021-11-22"
	}
	return s
}

func (s *SlotFactory) Build() []*mysql.Slot {

	s.fill_default()

	slotFactory := factory.NewFactory(&mysql.Slot{}).
		Attr("Date", func(args factory.Args) (interface{}, error) {
			date := time.Now().AddDate(0, 0, randomdata.Number(1, 7))
			return &date, nil
		}).
		SeqInt("Position", func(n int) (interface{}, error) {
			pos := int32(n)
			return &pos, nil
		}).
		Attr("Cost", func(args factory.Args) (interface{}, error) {
			cost := randomdata.Decimal(1, 100, 2)
			return &cost, nil
		}).
		SeqString("Status", func(n string) (interface{}, error) {
			return &s.Status[randomdata.Number(len(s.Status))], nil
		}).
		Attr("BookedBy", func(args factory.Args) (interface{}, error) {
			slot := args.Instance().(*mysql.Slot)
			if *slot.Status == models.SlotStatusBooked {
				uid := uuid.New().String()
				return &uid, nil
			}
			return nil, nil
		}).
		Attr("BookedDate", func(args factory.Args) (interface{}, error) {
			slot := args.Instance().(*mysql.Slot)
			if *slot.Status == models.SlotStatusBooked {
				date := slot.Date.AddDate(0, 0, 1)
				return &date, nil
			}
			return nil, nil
		})
	var slots []*mysql.Slot
	for i := 1; i <= s.Instances; i++ {
		slot := slotFactory.MustCreate().(*mysql.Slot)
		pos := int32(i)
		slot.Position = &pos
		slots = append(slots, slot)
	}
	return slots
}

type TransactionFactory struct {
	mysql.Transaction
	Instances int
}

func (t *TransactionFactory) WithId(id string) *TransactionFactory {
	t.Txnid = id
	return t
}

func (t *TransactionFactory) WithDate(date *time.Time) *TransactionFactory {
	t.Date = date
	return t
}

func (t *TransactionFactory) WithPosition(position *int32) *TransactionFactory {
	t.Position = position
	return t
}

func (t *TransactionFactory) WithInstances(n int) *TransactionFactory {
	t.Instances = n
	return t
}

func (t *TransactionFactory) Build() []*mysql.Transaction {
	txnFactory := factory.NewFactory(&mysql.Transaction{}).
		Attr("Date", func(args factory.Args) (interface{}, error) {
			if t.Date != nil {
				return t.Date, nil
			}
			date := time.Now().AddDate(0, 0, randomdata.Number(1, 7))
			return &date, nil
		}).
		SeqInt("Position", func(n int) (interface{}, error) {
			if t.Position != nil {
				return t.Position, nil
			}
			pos := int32(n)
			return &pos, nil
		}).
		Attr("Txnid", func(args factory.Args) (interface{}, error) {
			if t.Txnid != "" {
				return t.Txnid, nil
			}
			return uuid.New().String(), nil
		})
	var records []*mysql.Transaction
	for i := 1; i <= t.Instances; i++ {
		records = append(records, txnFactory.MustCreate().(*mysql.Transaction))
	}
	return records
}

type TestCreateSlotRequestBodyFactory struct {
	Instances int
	StartDate *models.JSONDate
	EndDate   *models.JSONDate
	Position  []int32
}

func (t *TestCreateSlotRequestBodyFactory) WithPositionRange(start, end int32) *TestCreateSlotRequestBodyFactory {
	t.Position = []int32{start, end}
	return t
}

func (t *TestCreateSlotRequestBodyFactory) WithDateRange(start, end time.Time) *TestCreateSlotRequestBodyFactory {
	st, en := models.JSONDate(start), models.JSONDate(end)
	t.StartDate, t.EndDate = &st, &en
	return t
}

func (t *TestCreateSlotRequestBodyFactory) WithInstances(n int) *TestCreateSlotRequestBodyFactory {
	t.Instances = n
	return t
}

func (t *TestCreateSlotRequestBodyFactory) Build() []*api.CreateSlotRequestBody {
	crbFactory := factory.NewFactory(&api.CreateSlotRequestBody{}).
		Attr("StartDate", func(args factory.Args) (interface{}, error) {
			if t.StartDate != nil {
				return *t.StartDate, nil
			}
			date := time.Now().AddDate(0, 0, randomdata.Number(1, 7))
			return models.JSONDate(date), nil
		}).
		Attr("EndDate", func(args factory.Args) (interface{}, error) {
			if t.EndDate != nil {
				return *t.EndDate, nil
			}
			date := time.Time(args.Instance().(*api.CreateSlotRequestBody).StartDate)
			date = date.AddDate(0, 0, randomdata.Number(0, 7))
			return models.JSONDate(date), nil
		}).
		SeqInt("Position", func(n int) (interface{}, error) {
			if t.Position != nil {
				return t.Position, nil
			}
			pos := int32(randomdata.Number(1, 10))
			return []int32{1, pos}, nil
		}).
		Attr("Cost", func(args factory.Args) (interface{}, error) {
			cost := randomdata.Decimal(100)
			return &cost, nil
		})
	var records []*api.CreateSlotRequestBody
	for i := 1; i <= t.Instances; i++ {
		records = append(records, crbFactory.MustCreate().(*api.CreateSlotRequestBody))
	}
	return records
}

type TestReserveSlotFactory struct {
	Date      *models.JSONDate
	Position  int32
	Instances int
}

func (t *TestReserveSlotFactory) WithDate(d time.Time) *TestReserveSlotFactory {
	date := models.JSONDate(d)
	t.Date = &date
	return t
}

func (t *TestReserveSlotFactory) WithPosition(i int32) *TestReserveSlotFactory {
	t.Position = i
	return t
}

func (t *TestReserveSlotFactory) WithInstances(n int) *TestReserveSlotFactory {
	t.Instances = n
	return t
}

func (t *TestReserveSlotFactory) Build() []*api.ReserveSlotRequestBody {
	rrbFactory := factory.NewFactory(&api.ReserveSlotRequestBody{}).
		Attr("Date", func(args factory.Args) (interface{}, error) {
			if t.Date != nil {
				return *t.Date, nil
			}
			date := time.Now().AddDate(0, 0, randomdata.Number(1, 7))
			return models.JSONDate(date), nil
		}).
		SeqInt("Position", func(n int) (interface{}, error) {
			if t.Position > 0 {
				return &t.Position, nil
			}
			pos := int32(n)
			return &pos, nil
		})
	var records []*api.ReserveSlotRequestBody
	for i := 1; i <= t.Instances; i++ {
		records = append(records, rrbFactory.MustCreate().(*api.ReserveSlotRequestBody))
	}
	return records
}

func TestFactory(t *testing.T) {
	factory := SlotFactory{}
	slots := factory.WithInstances(10).Build()
	assert.NotNil(t, slots)
	txnFactory := TransactionFactory{}
	txn := txnFactory.WithInstances(10).Build()
	assert.NotNil(t, txn)
	crbFactory := TestCreateSlotRequestBodyFactory{}
	crb := crbFactory.WithInstances(10).Build()
	assert.NotNil(t, crb)
	rrbFactory := TestReserveSlotFactory{}
	rrb := rrbFactory.WithInstances(10).Build()
	assert.NotNil(t, rrb)
}

type TestTemplateObject struct {
	Search  []Search  `json:"search" json:"search"`
	Create  []Create  `yaml:"create" json:"create"`
	Update  []Create  `yaml:"update" json:"update"`
	Reserve []Reserve `yaml:"reserve" json:"reserve"`
}

type Create struct {
	Description        string             `yaml:"description" json:"description"`
	Before             *Before            `yaml:"before,omitempty" json:"before"`
	Request            []CreateRequest    `yaml:"request" json:"request"`
	After              *After             `yaml:"after,omitempty" json:"after"`
	TestRequiredParams TestRequiredParams `yaml:"params" json:"params"`
}

type Search struct {
	Description        string             `yaml:"description" json:"description"`
	Before             *Before            `yaml:"before,omitempty" json:"before"`
	After              *After             `yaml:"after,omitempty" json:"after"`
	Query              Query              `yaml:"query" json:"query"`
	TestRequiredParams TestRequiredParams `yaml:"params" json:"params"`
}

type Reserve struct {
	Description        string             `yaml:"description" json:"description"`
	Before             []*ReserveBefore   `yaml:"before,omitempty" json:"before"`
	After              []*After           `yaml:"after,omitempty" json:"after"`
	Request            []ReserveRequest   `yaml:"request" json:"request"`
	TestRequiredParams TestRequiredParams `yaml:"params" json:"params"`
}

type Query struct {
	StartDate *Date   `yaml:"start_date" json:"start_date,omitempty"`
	EndDate   *Date   `yaml:"end_date" json:"end_date,omitempty"`
	Position  *int    `yaml:"position" json:"position,omitempty"`
	Status    *string `yaml:"status" json:"status,omitempty"`
	Uid       *string `yaml:"uid" json:"uid,omitempty"`
}

type After struct {
	Request            []DeleteRequest    `yaml:"request" json:"request"`
	TestRequiredParams TestRequiredParams `yaml:"params" json:"params"`
}
type Before struct {
	Request            []CreateRequest    `yaml:"request" json:"request"`
	TestRequiredParams TestRequiredParams `yaml:"params" json:"params"`
}
type ReserveBefore struct {
	Request            []interface{}      `yaml:"request" json:"request"`
	TestRequiredParams TestRequiredParams `yaml:"params" json:"params"`
}

type DeleteRequest struct {
	StartDate Date  `yaml:"start_date,omitempty" json:"start_date,omitempty"`
	EndDate   Date  `yaml:"end_date,omitempty" json:"end_date,omitempty"`
	Position  []int `yaml:"position,omitempty" json:"position,omitempty"`
}
type CreateRequest struct {
	StartDate *Date    `yaml:"start_date,omitempty" json:"start_date,omitempty"`
	EndDate   *Date    `yaml:"end_date,omitempty" json:"end_date,omitempty"`
	Position  []int    `yaml:"position,omitempty" json:"position,omitempty"`
	Cost      *float64 `yaml:"cost,omitempty" json:"cost,omitempty"`
}

type ReserveRequest struct {
	Date     *Date `yaml:"date" json:"date"`
	Position int   `yaml:"position" json:"position"`
}

type TestRequiredParams struct {
	ExpectedStatus int            `yaml:"expected_status" json:"expected_status"`
	ExpectedError  bool           `yaml:"expected_error" json:"expected_error"`
	ExpectedOutput ExpectedOutput `yaml:"expected_output,omitempty" json:"expected_output"`
	Url            string         `yaml:"url" json:"url"`
	Method         string         `yaml:"method" json:"method"`
}

type ExpectedOutput struct {
	Empty  bool        `yaml:"empty" json:"empty"`
	Output interface{} `yaml:"output" json:"output"`
}

type Date string

func (t *Date) UnmarshalYAML(unmarshal func(interface{}) error) error {

	var buf string
	err := unmarshal(&buf)
	if err != nil {
		return err
	}
	switch {
	case buf == "now":
		*t = Date(time.Now().Format(time.DateOnly))
	case strings.HasPrefix(buf, "now+"):
		ns := strings.Split(buf, "+")[1]
		days, err := strconv.Atoi(ns)
		if err != nil {
			return fmt.Errorf("invalid number of days to add after %s", buf)
		}
		*t = Date(time.Now().AddDate(0, 0, days).Format(time.DateOnly))
	case strings.HasPrefix(buf, "now-"):
		ns := strings.Split(buf, "-")[1]
		days, err := strconv.Atoi(ns)
		if err != nil {
			return fmt.Errorf("invalid number of days to add after %s", buf)
		}
		*t = Date(time.Now().AddDate(0, 0, -1*days).Format(time.DateOnly))
	}
	return nil
}

func (t Date) MarshalYAML() (interface{}, error) {
	return string(t), nil
}
