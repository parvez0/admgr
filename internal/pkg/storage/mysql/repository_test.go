package mysql_test

import (
	"fmt"
	"github.com/kiran-anand14/admgr/internal/pkg/models"
	"github.com/kiran-anand14/admgr/internal/pkg/storage/mysql"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"testing"
	"time"
)

type RepositoryTestSuite struct {
	suite.Suite
	repository *mysql.Storage
}

func (r *RepositoryTestSuite) BeforeTest(suiteName, test string) {
	conf := &models.DBConf{
		Host:     "localhost",
		Port:     "3306",
		Name:     "test_db",
		Username: "root",
		Password: "password",
	}
	logrus.SetLevel(logrus.DebugLevel)
	logger := logrus.New()
	var err error
	r.repository, err = mysql.NewStorage(logger, conf)
	assert.Nil(r.T(), err, fmt.Sprintf("MysqlSeedingFailed::%+v", conf))
}
func (r *RepositoryTestSuite) AfterTest(suiteName, test string) {
	if r.repository == nil {
		r.T().Fatalf("DB instance not initialized")
	}
	err := r.repository.DropAll()
	assert.Nil(r.T(), err, "Failed to drop tables")
}

func (r *RepositoryTestSuite) Test_Create() {
	slotFactory, transFactory := SlotFactory{}, TransactionFactory{}
	slots := slotFactory.
		WithStatus([]string{models.SlotStatusOpen}).
		WithInstances(10).
		Build()
	var transactions []*mysql.Transaction
	for _, slot := range slots {
		txn := transFactory.
			WithDate(slot.Date).
			WithPosition(slot.Position).
			WithInstances(1).
			Build()
		transactions = append(transactions, txn...)
	}
	slotAffected, slotErr := r.repository.Create(slots)
	txnAffected, txnErr := r.repository.Create(transactions)
	// Testing insert multiple records
	assert.Nil(r.T(), slotErr, "Failed to create slots")
	assert.Nil(r.T(), txnErr, "Failed to create transaction")
	assert.Equal(r.T(), len(slots), slotAffected, fmt.Sprintf("Expected to create %d slot records", len(slots)))
	assert.Equal(r.T(), len(transactions), txnAffected, fmt.Sprintf("Expected to create %d transactions records", len(transactions)))
	// Testing insert duplicate entries
	_, slotErr = r.repository.Create(slots)
	_, txnErr = r.repository.Create(transactions)
	assert.Error(r.T(), slotErr, "Expected a slot duplication")
	assert.Error(r.T(), txnErr, "Expected a transaction duplication")
}

func (r *RepositoryTestSuite) Test_Create_NullValue() {
	// Testing insert null value entries
	slotF, txnF := SlotFactory{}, TransactionFactory{}
	nullValSlot := slotF.WithInstances(1).Build()[0]
	nullValTxn := txnF.WithInstances(1).Build()[0]
	nullValTxn.Date = nil
	nullValSlot.Cost = nil
	_, slotErr := r.repository.Create(nullValSlot)
	_, txnErr := r.repository.Create(nullValTxn)
	assert.Error(r.T(), slotErr, "Expected a slot duplication")
	assert.Error(r.T(), txnErr, "Expected a transaction duplication")
}

func (r *RepositoryTestSuite) Test_Update() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	slotF := SlotFactory{}

	slots := slotF.WithStatus([]string{models.SlotStatusOpen}).WithInstances(5).Build()
	_, err := r.repository.Create(slots)
	assert.Nil(r.T(), err, "Failed to create slots")
	for i := range slots {
		status := []string{models.SlotStatusClosed, models.SlotStatusBooked, models.SlotStatusHold}[rand.Intn(3)]
		slots[i].Status = &status
	}
	affected, err := r.repository.UpdateSlots(slots)
	assert.Nil(r.T(), err, "Expected to update slot status to booked")
	assert.Equal(r.T(), len(slots), affected)

	// Test for failed update and revert
	date := time.Now().AddDate(0, 0, -1)
	slots[2].Date = &date
	for i := range slots {
		status := models.SlotStatusOpen
		slots[i].Status = &status
	}
	affected, err = r.repository.UpdateSlots(slots)
	assert.NotNil(r.T(), err, "Expected to update slot status to booked")
	assert.Equal(r.T(), 0, affected)

}

func (r *RepositoryTestSuite) Test_Search() {
	slotF := SlotFactory{}
	slot := slotF.WithInstances(1).Build()[0]
	_, err := r.repository.Create(slot)
	assert.Nil(r.T(), err, "Expected to create slots")
	filters := map[string]interface{}{
		"date":     slot.Date.Format(time.DateOnly),
		"position": slot.Position,
	}
	slotRes, err := r.repository.SearchSlots(filters)
	assert.Nil(r.T(), err)
	if err == nil {
		assert.Equal(r.T(), slotRes[0].Date.Format(time.DateOnly), slot.Date.Format(time.DateOnly), "Expected search result to be equal")
		assert.Equal(r.T(), slotRes[0].Position, slot.Position, "Expected search result to be equal")
		assert.Equal(r.T(), slotRes[0].Status, slot.Status, "Expected search result to be equal")
		assert.Equal(r.T(), slotRes[0].Cost, slot.Cost, "Expected search result to be equal")
	}
	filters = map[string]interface{}{
		"date": time.Now().Format(time.DateOnly),
	}
	slotRes, err = r.repository.SearchSlots(filters)
	assert.Nil(r.T(), err)
	assert.Empty(r.T(), slotRes)
}

func (r *RepositoryTestSuite) Test_SearchWithPrimaryKeyAndStatus() {
	slotF := SlotFactory{}
	slots := slotF.WithInstances(5).Build()
	_, err := r.repository.Create(slots)
	assert.Nil(r.T(), err, "Expected to create slots")
	for _, s := range slots {
		firstSlot, err := r.repository.SearchSlotsByPrimaryKeyAndStatus(*s.Date, *s.Position, *s.Status)
		assert.Nil(r.T(), err)
		assert.Equal(r.T(), *s.Status, *firstSlot.Status)
		assert.Equal(r.T(), *s.Cost, *firstSlot.Cost)
		assert.Equal(r.T(), s.Date.Format(time.DateOnly), firstSlot.Date.Format(time.DateOnly))
		assert.Equal(r.T(), *s.Position, *firstSlot.Position)
	}
}

func (r *RepositoryTestSuite) Test_UpdateSlotsStatus() {
	slotF := SlotFactory{}
	slot := slotF.WithStatus([]string{models.SlotStatusOpen}).WithInstances(1).Build()
	_, err := r.repository.Create(slot)
	assert.Nil(r.T(), err, "Expected to create slots")
	err = r.repository.UpdateSlotsStatus(slot, *slot[0].Status, models.SlotStatusClosed)
	assert.Nil(r.T(), err)
	date, position, status := *slot[0].Date, *slot[0].Position, models.SlotStatusClosed
	firstSlot, err := r.repository.SearchSlotsByPrimaryKeyAndStatus(date, position, status)
	assert.Nil(r.T(), err)
	assert.NotNil(r.T(), firstSlot)
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}
