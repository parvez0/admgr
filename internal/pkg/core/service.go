package core

import (
	"github.com/kiran-anand14/admgr/internal/pkg/api"
	"github.com/kiran-anand14/admgr/internal/pkg/models"
	"github.com/sirupsen/logrus"
)

// Service provides User adding operations.
type Service interface {
	CreateProd(prod *api.Product) (map[string]interface{}, *models.Error)
}

// Repository provides access to User repository.
type Repository interface {
	CreateProd(prod *api.Product) (string, *models.Error)
}

type service struct {
	log *logrus.Logger
	r   Repository
}

// NewService creates an adding service with the necessary dependencies
func NewService(r Repository, log *logrus.Logger) Service {
	s := service{
		log: log,
		r:   r,
	}

	return &s
}

func (s *service) CreateProd(prod *api.Product) (map[string]interface{}, *models.Error) {
	// any validation can be done here

	s.log.Infof("Create Product: Adding %v to Repository", *prod)

	id, er := s.r.CreateProd(prod)
	if er != nil {
		return nil, er
	}

	m := make(map[string]interface{})
	m["id"] = id

	return m, nil
}
