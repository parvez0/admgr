package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kiran-anand14/admgr/internal/pkg/api"
	"github.com/kiran-anand14/admgr/internal/pkg/models"
	"github.com/kiran-anand14/admgr/internal/pkg/storage/mysql"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const ContentTypeJSON = "application/json"

type AccountingService interface {
	Debit(slots []*mysql.Slot, uid, txnid string) error
}

type accountingService struct {
	url        string
	source     string
	log        *logrus.Logger
	restClient *http.Client
}

func (a accountingService) Debit(slots []*mysql.Slot, uid, txnid string) error {
	var metaSlots []api.AccountingMetadataSlot
	var totalAmount float64
	for _, s := range slots {
		metaSlots = append(metaSlots, api.AccountingMetadataSlot{
			Date:     *s.Date,
			Position: *s.Position,
			Cost:     *s.Cost,
		})
		totalAmount += *s.Cost
	}
	accountRequest := api.AccountingRequestBody{
		Source: a.source,
		Uid:    uid,
		Amount: totalAmount,
		Txnid:  txnid,
		Metadata: api.AccountingMetadata{
			Slots: metaSlots,
		},
	}
	a.log.Debugf("Initiating debit transaction: %+v", accountRequest)
	jsonPayload, err := json.Marshal(accountRequest)
	if err != nil {
		return models.NewError(
			fmt.Sprintf("JSON marshal failed for accounting transation [Error: %s, Object:%+v]", err.Error(), accountRequest),
			models.DecodeFailureError,
		)
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/debit", a.url), bytes.NewBuffer(jsonPayload))
	if err != nil {
		return models.NewError(
			fmt.Sprintf("RestRequestFormation failed %s", err.Error()),
			models.DecodeFailureError,
		)
	}
	res, err := a.restClient.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		statusCode := -1
		if res != nil {
			statusCode = res.StatusCode
		}
		a.log.Errorf("DebitTransactionFailed::[StatusCode: %d, Error: %v]", statusCode, err)
		return models.NewError(
			"Debit transaction failed",
			models.DependentServiceRequestFailed,
		)
	}
	return nil
}

func NewAccountingService(host, port, source string, logger *logrus.Logger) AccountingService {
	return accountingService{
		url:    fmt.Sprintf("%s:%s", host, port),
		log:    logger,
		source: source,
		restClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}
}
