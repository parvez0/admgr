package mysql

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/kiran-anand14/admgr/internal/pkg/models"
	"gorm.io/gorm"
	"time"
)

// Slot represents a slot in the ad manager system.
type Slot struct {
	Date        *time.Time   `gorm:"primaryKey;type:date;not null" json:"date"`
	Position    *int32       `gorm:"primaryKey;type:int;not null" json:"position"`
	Cost        *float64     `gorm:"type:decimal(10,2);not null" json:"cost"`
	Status      *string      `gorm:"type:varchar(45);not null" json:"status"`
	Created     time.Time    `gorm:"default:CURRENT_TIMESTAMP" json:"created"`
	Modified    time.Time    `gorm:"autoUpdateTime" json:"modified"`
	BookedDate  *time.Time   `gorm:"type:datetime" json:"booked_date,omitempty"`
	BookedBy    *string      `gorm:"type:varchar(36)" json:"booked_by,omitempty"`
	Transaction *Transaction `gorm:"ForeignKey:Date,Position;References:Date,Position;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (s *Slot) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

func (s *Slot) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	if b, ok := value.([]byte); ok {
		return json.Unmarshal(b, s)
	}
	return errors.New("failed to scan SlotJSON")
}

func (s *Slot) ToString() string {
	bs, _ := json.Marshal(s)
	return string(bs)
}

// Transaction represents a transaction in the ad manager system.
type Transaction struct {
	Txnid    string     `gorm:"type:varchar(36)" json:"txnid"`
	Created  time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created"`
	Date     *time.Time `gorm:"primaryKey;type:date;not null" json:"date"`
	Position *int32     `gorm:"primaryKey;type:int;not null" json:"position"`
}

// TableName Define foreign key relationship
func (t *Transaction) TableName() string {
	return "transactions"
}

func (t *Transaction) AfterCreate(tx *gorm.DB) (err error) {
	var slot Slot
	if t.Date == nil {
		return models.NewError("column 'date' cannot be empty", models.ActionForbidden)
	}
	if err = tx.Model(&Slot{}).Where(
		"date = ? AND position = ? AND status = ?",
		t.Date.Format(time.DateOnly),
		t.Position,
		models.SlotStatusOpen).
		Update("status", models.SlotStatusHold).Error; err != nil {
		return err
	}
	// Set foreign key
	t.Date = slot.Date
	t.Position = slot.Position
	return nil
}

func (t *Transaction) AfterDelete(tx *gorm.DB) (err error) {
	var slot Slot
	if t.Date == nil {
		return models.NewError("column 'date' cannot be empty", models.ActionForbidden)
	}
	if err = tx.Model(&Slot{}).Where(
		"date = ? AND position = ? AND status = ?",
		t.Date.Format(time.DateOnly),
		t.Position,
		models.SlotStatusHold).
		Update("status", models.SlotStatusOpen).Error; err != nil {
		return err
	}
	// Set foreign key
	t.Date = slot.Date
	t.Position = slot.Position
	return nil
}

type GetOptions struct {
	StartDate          time.Time
	EndDate            time.Time
	PositionStart      string
	PositionEnd        string
	Status             string
	Uid                string
	Query              string
	PreloadTransaction bool
}
