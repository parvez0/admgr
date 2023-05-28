package tests_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kiran-anand14/admgr/internal/pkg/core"
	"github.com/kiran-anand14/admgr/internal/pkg/http/rest"
	"github.com/kiran-anand14/admgr/internal/pkg/models"
	"github.com/kiran-anand14/admgr/internal/pkg/storage/mysql"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	UrlHealthCheck    = "health-check"
	UrlAdslotsRequest = "adslots"
	UrlReserveRequest = "adslots/reserve"
	ContentTypeJson   = "application/json"
)

type Address struct {
	Host string
	Port string
}

func (a *Address) Url() string {
	return fmt.Sprintf("http://%s:%s", a.Host, a.Port)
}

type HttRestTestSuite struct {
	suite.Suite
	repository *mysql.Storage
	url        string
	ctxCancel  context.CancelFunc
	ch         chan bool
	template   *TestTemplateObject
	client     *http.Client
}

func (r *HttRestTestSuite) BeforeTest(suiteName, test string) {
	writer := io.MultiWriter(os.Stdout)
	account := Address{"http://localhost", "10002"}
	admgr := Address{"localhost", "10001"}
	logger := &logrus.Logger{}
	dbConf := models.DBConf{
		Host:     "localhost",
		Port:     "3306",
		Name:     "test_db",
		Username: "root",
		Password: "password",
	}
	s, err := mysql.NewStorage(logger, writer, "error", &dbConf)
	if err != nil {
		logger.Errorf("%s", err.Error())
		return
	}

	accountService := core.NewAccountingService(account.Host, account.Port, "admgr", logger)
	service := core.NewService(s, accountService, logger)

	router, _ := rest.Handler(logger, service, os.Stdout)
	r.repository = s
	r.url = admgr.Url()

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", admgr.Host, admgr.Port),
		Handler: router,
	}

	go func() {
		// Listen and serve HTTP server
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			assert.Nil(r.T(), err)
		}
	}()
	go func() {
		quit := make(chan bool, 1)
		r.ch = quit
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Shutdown the server gracefully
		if err := server.Shutdown(ctx); err != nil {
			fmt.Println("Server forced to shutdown:", err)
		}
		r.ctxCancel = cancel
	}()

	r.client = &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   10 * time.Second,
	}

	buf, err := os.ReadFile("./test_data.yml")
	assert.Nil(r.T(), err, "Test template not found at './test_data.yml'")
	var testTemplate TestTemplateObject
	err = yaml.Unmarshal(buf, &testTemplate)
	assert.Nil(r.T(), err, "Failed to parse template object")
	r.template = &testTemplate
}

func (r *HttRestTestSuite) AfterTest(suiteName, test string) {
	if r.repository == nil {
		r.T().Fatalf("DB instance not initialized")
	}
	err := r.repository.DropAll()
	assert.Nil(r.T(), err, "Failed to drop tables")
	r.ch <- true
}

func (r *HttRestTestSuite) Test_ServerWorking() {
	resp, err := http.Get(join(r.url, UrlHealthCheck))
	assert.Nil(r.T(), err)
	if err != nil {
		assert.Equal(r.T(), http.StatusOK, resp.StatusCode)
	}
	r.ctxCancel()
}

func (r *HttRestTestSuite) TestCreateSlot() {
	for _, test := range r.template.Create {
		r.T().Run(test.Description, func(t *testing.T) {
			// Perform any necessary setup
			if test.Before != nil {
				req := createRequest(join(r.url, test.Before.TestRequiredParams.Url), test.Before.TestRequiredParams.Method, test.Before.Request, nil, false)
				res, err := r.client.Do(req)
				assertError(t, err, test.Before.TestRequiredParams)
				checkResponse(t, res, test.Before.TestRequiredParams)
			}

			// Send the test request
			req := createRequest(join(r.url, test.TestRequiredParams.Url), test.TestRequiredParams.Method, test.Request, nil, false)
			res, err := r.client.Do(req)
			assertError(t, err, test.TestRequiredParams)
			checkResponse(t, res, test.TestRequiredParams)

			// Perform any necessary cleanup
			if test.After != nil {
				req := createRequest(join(r.url, test.After.TestRequiredParams.Url), test.After.TestRequiredParams.Method, test.After.Request, nil, false)
				res, err := r.client.Do(req)
				assertError(t, err, test.After.TestRequiredParams)
				checkResponse(t, res, test.After.TestRequiredParams)
			}
		})
	}
}

func (r *HttRestTestSuite) TestUpdateSlot() {
	for _, test := range r.template.Update {
		r.T().Run(test.Description, func(t *testing.T) {
			// Perform any necessary setup
			if test.Before != nil {
				req := createRequest(join(r.url, test.Before.TestRequiredParams.Url), test.Before.TestRequiredParams.Method, test.Before.Request, nil, false)
				res, err := r.client.Do(req)
				assertError(t, err, test.Before.TestRequiredParams)
				checkResponse(t, res, test.Before.TestRequiredParams)
			}

			// Send the test request
			req := createRequest(join(r.url, test.TestRequiredParams.Url), test.TestRequiredParams.Method, test.Request, nil, false)
			res, err := r.client.Do(req)
			assertError(t, err, test.TestRequiredParams)
			checkResponse(t, res, test.TestRequiredParams)

			// Perform any necessary cleanup
			if test.After != nil {
				req := createRequest(join(r.url, test.After.TestRequiredParams.Url), test.After.TestRequiredParams.Method, test.After.Request, nil, false)
				res, err := r.client.Do(req)
				assertError(t, err, test.After.TestRequiredParams)
				checkResponse(t, res, test.After.TestRequiredParams)
			}
		})
	}
}

func (r *HttRestTestSuite) TestSearchSlot() {
	for _, test := range r.template.Search {
		r.T().Run(test.Description, func(t *testing.T) {
			// Perform any necessary setup
			if test.Before != nil {
				req := createRequest(join(r.url, test.Before.TestRequiredParams.Url), test.Before.TestRequiredParams.Method, test.Before.Request, nil, false)
				res, err := r.client.Do(req)
				assertError(t, err, test.Before.TestRequiredParams)
				checkResponse(t, res, test.Before.TestRequiredParams)
			}

			// Send the test request
			req := createRequest(join(r.url, test.TestRequiredParams.Url), test.TestRequiredParams.Method, nil, test.Query, false)
			res, err := r.client.Do(req)
			assertError(t, err, test.TestRequiredParams)
			checkResponse(t, res, test.TestRequiredParams)

			// Perform any necessary cleanup
			if test.After != nil {
				req := createRequest(join(r.url, test.After.TestRequiredParams.Url), test.After.TestRequiredParams.Method, test.After.Request, nil, false)
				res, err := r.client.Do(req)
				assertError(t, err, test.After.TestRequiredParams)
				checkResponse(t, res, test.After.TestRequiredParams)
			}
		})
	}
}

func (r *HttRestTestSuite) TestReserveSlot() {
	for _, test := range r.template.Reserve {
		r.T().Run(test.Description, func(t *testing.T) {
			// Perform any necessary setup
			for _, beforeTest := range test.Before {
				req := createRequest(join(r.url, beforeTest.TestRequiredParams.Url), beforeTest.TestRequiredParams.Method, beforeTest.Request, nil, true)
				res, err := r.client.Do(req)
				assertError(t, err, beforeTest.TestRequiredParams)
				checkResponse(t, res, beforeTest.TestRequiredParams)
			}

			// Send the test request
			req := createRequest(join(r.url, test.TestRequiredParams.Url), test.TestRequiredParams.Method, test.Request, test.Query, false)
			res, err := r.client.Do(req)
			assertError(t, err, test.TestRequiredParams)
			checkResponse(t, res, test.TestRequiredParams)

			// Perform any necessary cleanup
			for _, afterTest := range test.After {
				req := createRequest(join(r.url, afterTest.TestRequiredParams.Url), afterTest.TestRequiredParams.Method, afterTest.Request, nil, false)
				res, err := r.client.Do(req)
				assertError(t, err, afterTest.TestRequiredParams)
				checkResponse(t, res, afterTest.TestRequiredParams)
			}
			r.repository.DropAll()
			r.repository.Initialize()
		})
	}
}

func assertError(t *testing.T, err error, test TestRequiredParams) {
	if test.ExpectedError {
		assert.Error(t, err)
	} else {
		assert.Nil(t, err)
	}
}

func createRequest(url string, method string, requests interface{}, query interface{}, decode bool) *http.Request {
	if decode {
		buf, _ := json.Marshal(requests)
		bufStr := parseDateFromTemplate(string(buf))
		json.Unmarshal([]byte(bufStr), &requests)
	}
	switch strings.ToUpper(method) {
	case http.MethodGet:
		req, _ := http.NewRequest(method, url, nil)
		q := req.URL.Query()
		var params map[string]interface{}
		buf, _ := json.Marshal(query)
		json.Unmarshal(buf, &params)
		for k, v := range params {
			q.Add(k, fmt.Sprintf("%v", v))
		}
		req.URL.RawQuery = q.Encode()
		return req
	default:
		var body bytes.Buffer
		json.NewEncoder(&body).Encode(requests)
		req, _ := http.NewRequest(method, url, &body)
		req.Header.Set("Content-Type", ContentTypeJson)
		return req
	}
	return nil
}

func checkResponse(t *testing.T, rr *http.Response, test TestRequiredParams) {
	// Check the response status code
	if rr == nil {
		if test.ExpectedError {
			return
		}
		assert.NotNil(t, rr, "Expected response to be not nil")
		return
	}
	if status := rr.StatusCode; status != test.ExpectedStatus {
		t.Errorf("ExpectedStatusCode: %v, Got %v : Error: %+v",
			test.ExpectedStatus, status, readIo(rr.Body))
	}

	if test.ExpectedOutput.Empty || test.ExpectedOutput.Output == nil {
		return
	}
	body, _ := json.Marshal(test.ExpectedOutput.Output)
	expectedResponseOutput := parseDateFromTemplate(string(body))
	require.JSONEq(t, expectedResponseOutput, readIo(rr.Body))
}

func readIo(body io.ReadCloser) string {
	//Check the response body
	tmpBytes, _ := io.ReadAll(body)
	var tmpBody []map[string]interface{}
	err := json.Unmarshal(tmpBytes, &tmpBody)
	if err == nil && len(tmpBody) > 0 {
		if _, e := tmpBody[0]["date"]; !e {
			return string(tmpBytes)
		}
		sort.Slice(tmpBody, func(i, j int) bool {
			x, _ := time.Parse(time.DateOnly, tmpBody[i]["date"].(string))
			y, _ := time.Parse(time.DateOnly, tmpBody[j]["date"].(string))
			if x.Before(y) {
				return true
			}
			return false
		})
		tmpBytes, _ = json.Marshal(tmpBody)
	}
	return string(tmpBytes)
}

func join(url, path string) string {
	return fmt.Sprintf("%s/%s", url, path)
}

func parseDateFromTemplate(b string) string {
	re := regexp.MustCompile(`now([+-]\d+)?`)
	for {
		matches := re.FindAllStringSubmatch(b, 1)
		if len(matches) == 0 {
			return b
		}
		dateString := matches[0][0]
		switch {
		case strings.HasPrefix(dateString, "now+"):
			strNum, _ := strconv.Atoi(strings.Split(dateString, "+")[1])
			b = strings.Replace(b, dateString, time.Now().AddDate(0, 0, strNum).Format(time.DateOnly), 1)
		case strings.HasPrefix(dateString, "now-"):
			strNum, _ := strconv.Atoi(strings.Split(dateString, "+")[1])
			b = strings.Replace(b, dateString, time.Now().AddDate(0, 0, -1*strNum).Format(time.DateOnly), 1)
		default:
			b = strings.Replace(b, "now", time.Now().Format(time.DateOnly), 1)
		}
	}
	return b
}

func TestHttRestTestSuite(t *testing.T) {
	suite.Run(t, new(HttRestTestSuite))
}
