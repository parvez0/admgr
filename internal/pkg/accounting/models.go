package accounting

import "time"

type AccountingRequestBody struct {
	Source   string             `json:"source"`
	Uid      string             `json:"uid"`
	Amount   float64            `json:"amount"`
	Txnid    string             `json:"txnid"`
	Metadata AccountingMetadata `json:"metadata"`
}

type AccountingMetadata struct {
	Slots []AccountingMetadataSlot `json:"slots"`
}

type AccountingMetadataSlot struct {
	Date     time.Time `json:"date"`
	Position int32     `json:"position"`
	Cost     float64   `json:"cost"`
}

type AccountingStatusResponse struct {
	Txnid    string             `json:"txnid,omitempty"`
	UID      string             `json:"uid,omitempty"`
	Created  time.Time          `json:"created,omitempty"`
	Metadata AccountingMetadata `json:"metadata,omitempty"`
}
