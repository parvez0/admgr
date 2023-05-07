package api

import (
	"github.com/kiran-anand14/admgr/internal/pkg/models"
	"time"
)

// All the structure definitions for decoding REST API payload can go here.

type ErrorResponse struct {
	Error string `json:"error"`
}

type CreateSlotRequestBody struct {
	StartDate models.JSONDate `json:"start_date,omitempty" binding:"required,date" validate:"json_date"`
	EndDate   models.JSONDate `json:"end_date,omitempty" binding:"required,date,gtefield=StartDate" validate:"json_date"`
	Position  []int32         `json:"position,omitempty" binding:"required" validate:"range"`
	Cost      *float64        `json:"cost,omitempty" binding:"required,min=0" validate:"required"`
	Status    *string         `json:"status,omitempty" binding:"required,oneof=open closed" validate:"required"`
}

type GetRequestParameters struct {
	StartDate models.JSONDate `json:"start_date,omitempty" form:"start_date" validate:"json_date"`
	EndDate   models.JSONDate `json:"end_date,omitempty" form:"end_date" validate:"json_date"`
	Position  *int32          `json:"position,omitempty" form:"position,omitempty"`
	Uid       *string         `json:"uid,omitempty" form:"uid,omitempty"`
	Status    *string         `json:"status,omitempty" form:"status,omitempty"`
}

type GetSlotsResponse struct {
	Date   time.Time       `json:"date"`
	Status string          `json:"status"`
	Slots  *[]SlotResponse `json:"slots,omitempty"`
}

type SlotResponse struct {
	Position   int32           `json:"position"`
	Cost       float64         `json:"cost"`
	Status     string          `json:"status"`
	BookedBy   *string         `json:"booked_by,omitempty"`
	BookedDate models.JSONDate `json:"booked_date,omitempty"`
}
