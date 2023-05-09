package models

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	SlotStatusOpen   = "open"
	SlotStatusClosed = "closed"
	SlotStatusBooked = "booked"
	SlotStatusHold   = "hold"
)

// JSONDate Custom time object with layout formatting
type JSONDate time.Time

func (jt *JSONDate) UnmarshalJSON(b []byte) error {
	t, err := time.ParseInLocation(`"`+time.DateOnly+`"`, string(b), time.Local)
	if err != nil {
		return err
	}
	*jt = JSONDate(t)
	return nil
}

func (jt *JSONDate) MarshalJSON() ([]byte, error) {
	t := time.Time(*jt)
	return json.Marshal(t.Format(time.DateOnly))
}

func PtrString(s string) *string {
	return &s
}

func PtrDate(d time.Time) *time.Time {
	return &d
}

func PtrInt(p int32) *int32 {
	return &p
}

func DateToString(d time.Time) string {
	return d.Format(time.DateOnly)
}

func JsonDate(d time.Time) JSONDate {
	return JsonDate(d)
}

func JsonDatePtr(d JSONDate) *JSONDate {
	return &d
}

func Int32ToString(d int32) string {
	return fmt.Sprintf("%d", d)
}
