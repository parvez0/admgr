package models

import (
	"encoding/json"
	"time"
)

const (
	SLOT_STATUS_OPEN   = "open"
	SLOT_STATUS_CLOSED = "closed"
	SLOT_STATUS_BOOKED = "booked"
	SLOT_STATUS_HOLD   = "hold"
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
