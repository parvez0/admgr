package tests_test

import (
	"github.com/kiran-anand14/admgr/internal/pkg/api"
	"github.com/kiran-anand14/admgr/internal/pkg/core"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ServiceTestSuite struct {
	suite.Suite
	service core.Service
}

func (s *ServiceTestSuite) BeforeTest() {
	r, a := NewMockRepository(), NewMockAccounting()
	s.service = core.NewService(r, a, &logrus.Logger{})
}

func (s *ServiceTestSuite) Test_CreateSlots() {
	var slots []*api.CreateSlotRequestBody
	s.service.CreateSlots(slots)
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
