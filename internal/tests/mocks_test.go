package tests_test

import (
	"github.com/kiran-anand14/admgr/internal/pkg/core"
	"github.com/kiran-anand14/admgr/internal/pkg/storage/mysql"
	"github.com/stretchr/testify/mock"
)

func NewMockRepository() core.Repository {
	return MockRepository{}
}

func NewMockAccounting() core.AccountingService {
	return MockAccounting{}
}

type MockRepository struct {
	mock.Mock
	ReturnUpdateValue int
	ReturnSlots       []*mysql.Slot
	ReturnError       error
}

func (m MockRepository) Create(i interface{}) (int, error) {
	return m.ReturnUpdateValue, m.ReturnError
}

func (m MockRepository) UpdateSlots(slots []*mysql.Slot) (int, error) {
	return m.ReturnUpdateValue, m.ReturnError
}

func (m MockRepository) SearchSlots(filters map[string]interface{}) ([]*mysql.Slot, error) {
	return m.ReturnSlots, m.ReturnError
}

func (m MockRepository) UpdateSlotsStatus(slots []*mysql.Slot, lastStatus, newStatus string) error {
	return m.ReturnError
}

type MockAccounting struct {
	mock.Mock
	ReturnError error
}

func (m MockAccounting) Debit(slots []*mysql.Slot, uid, txnid string) error {
	return m.ReturnError
}
