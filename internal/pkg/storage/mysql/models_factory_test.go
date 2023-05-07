package mysql_test

import (
	"github.com/Pallinder/go-randomdata"
	"github.com/bluele/factory-go/factory"
	"github.com/google/uuid"
	"github.com/kiran-anand14/admgr/internal/pkg/models"
	"github.com/kiran-anand14/admgr/internal/pkg/storage/mysql"
	"github.com/stretchr/testify/assert"
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
			models.SLOT_STATUS_OPEN,
			models.SLOT_STATUS_BOOKED,
			models.SLOT_STATUS_HOLD,
			models.SLOT_STATUS_CLOSED,
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
			if *slot.Status == models.SLOT_STATUS_BOOKED {
				uid := uuid.New().String()
				return &uid, nil
			}
			return nil, nil
		}).
		Attr("BookedDate", func(args factory.Args) (interface{}, error) {
			slot := args.Instance().(*mysql.Slot)
			if *slot.Status == models.SLOT_STATUS_BOOKED {
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

func TestFactory(t *testing.T) {
	factory := SlotFactory{}
	slots := factory.WithInstances(10).Build()
	assert.NotNil(t, slots)
	txnFactory := TransactionFactory{}
	txn := txnFactory.WithInstances(10).Build()
	assert.NotNil(t, txn)
}
