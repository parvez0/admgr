package accounting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kiran-anand14/admgr/internal/pkg/models"
	"github.com/kiran-anand14/admgr/internal/pkg/storage/mysql"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const ContentTypeJSON = "application/json"

type AccountingService interface {
	Debit(slots []*mysql.Slot, uid, txnid string) error
	Status(txnids []string) ([]*AccountingStatusResponse, error)
}

type accountingService struct {
	url        string
	source     string
	log        *logrus.Logger
	restClient *http.Client
}

func (a accountingService) Status(txnids []string) ([]*AccountingStatusResponse, error) {

	reqBody, _ := json.Marshal(txnids)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/status", a.url), bytes.NewReader(reqBody))
	if err != nil {
		return nil, models.NewError(
			fmt.Sprintf("RestRequestFormation failed %s", err.Error()),
			models.DecodeFailureError,
		)
	}
	a.log.Debugf("AccountingHandler: %s %s", req.Method, req.URL.String())
	res, err := a.restClient.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		statusCode := -1
		if res != nil {
			statusCode = res.StatusCode
		}
		a.log.Errorf("DebitTransactionFailed::[StatusCode: %d, Error: %v]", statusCode, err)
		return nil, models.NewError(
			"Debit transaction failed",
			models.DependentServiceRequestFailed,
		)
	}
	var statusResponse []*AccountingStatusResponse
	json.NewDecoder(res.Body).Decode(&statusResponse)
	return statusResponse, nil
}

func (a accountingService) Debit(slots []*mysql.Slot, uid, txnid string) error {
	var metaSlots []AccountingMetadataSlot
	var totalAmount float64
	for _, s := range slots {
		metaSlots = append(metaSlots, AccountingMetadataSlot{
			Date:     *s.Date,
			Position: *s.Position,
			Cost:     *s.Cost,
		})
		totalAmount += *s.Cost
	}
	accountRequest := AccountingRequestBody{
		Source: a.source,
		Uid:    uid,
		Amount: totalAmount,
		Txnid:  txnid,
		Metadata: AccountingMetadata{
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

func NewAccountingService(_log *logrus.Logger, conf models.AccountingServiceConf, source string) AccountingService {
	accService := accountingService{
		url:    fmt.Sprintf("%s://%s:%s", conf.Scheme, conf.Host, conf.Port),
		log:    _log,
		source: source,
		restClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}
	retries := 0
	var err *models.Error
	for {
		retries++
		healthCheckUrl := fmt.Sprintf("%s/%s", accService.url, conf.HealthCheckPath)
		res, err := http.Get(healthCheckUrl)
		if err == nil && res.StatusCode == http.StatusOK {
			_log.Infof("AccountingServiceInitialization:: service is active on %s, recieved acknowledgement", healthCheckUrl)
			break
		}
		if err == nil {
			err = models.NewError(
				fmt.Sprintf("Request failed on url %s with statusCode: %d", healthCheckUrl, res.StatusCode),
				models.DependentServiceRequestFailed)
		}
		_log.Errorf("AccountingServiceInitialization:: failed to check accounting service status, Error: %s, retrying %d", err, retries)
		if retries > 10 {
			break
		}
		time.Sleep(10 * time.Second)
	}
	if err != nil {
		_log.Fatalf("AccountingServiceInitialization:: Retries %d done for accounting service, but couldn't established the connection, Error: %s", retries, err)
	}
	return accService
}
