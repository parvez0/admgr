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
}

type ReserveSlotRequestBody struct {
	Date     models.JSONDate `json:"date,omitempty" validate:"json_date"`
	Position *int32          `json:"position" validate:"required"`
}

type GetSlotsResponse struct {
	Date   string          `json:"date"`
	Status string          `json:"status"`
	Slots  []*SlotResponse `json:"slots,omitempty"`
}

type SlotResponse struct {
	Position   int32            `json:"position"`
	Cost       float64          `json:"cost"`
	Status     string           `json:"status"`
	BookedBy   *string          `json:"booked_by,omitempty"`
	BookedDate *models.JSONDate `json:"booked_date,omitempty"`
}

type AccountingRequestBody struct {
	Source   string             `json:"source"`
	Uid      string             `json:"uid"`
	Amount   float64            `json:"amount"`
	Txnid    string             `json:"txnid"`
	Metadata AccountingMetadata `json:"metadata"`
}

type DeleteSlotRequestBody struct {
	StartDate models.JSONDate `json:"start_date,omitempty" binding:"required,date" validate:"json_date"`
	EndDate   models.JSONDate `json:"end_date,omitempty" binding:"required,date,gtefield=StartDate" validate:"json_date"`
	Position  []int32         `json:"position,omitempty" binding:"required" validate:"range"`
}

type AccountingMetadata struct {
	Slots []AccountingMetadataSlot `json:"slots"`
}

type AccountingMetadataSlot struct {
	Date     time.Time `json:"date"`
	Position int32     `json:"position"`
	Cost     float64   `json:"cost"`
}
