package mysql

import (
	"errors"
	"fmt"
	"time"

	"github.com/kiran-anand14/admgr/internal/pkg/api"
	"github.com/kiran-anand14/admgr/internal/pkg/storage"
	"github.com/sirupsen/logrus"

	sqlDrvMySql "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/kiran-anand14/admgr/internal/pkg/models"
)

type Storage struct {
	logger *logrus.Logger
	db     *gorm.DB
}

func NewStorage(log *logrus.Logger, dbConf *models.DBConf) (*Storage, error) {
	s := new(Storage)

	s.logger = log

	dsn := dbConf.Username + ":" + dbConf.Password + "@tcp" + "(" + dbConf.Host +
		":" + dbConf.Port + ")/" + dbConf.Name + "?" + "charset=utf8mb4&parseTime=True&loc=Local&clientFoundRows=true&timeout=60s"

	s.logger.Debugf("Database Connection String: %s", dsn)

	var db *gorm.DB
	var err error
	var retryCount uint8

	for {
		retryCount++
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			if retryCount == 60 {
				break
			}

			s.logger.Errorf("Error connecting to database : error=%v, retrying in 1s", err)
			time.Sleep(1 * time.Second)

			continue
		}
		break
	}
	if err != nil {
		s.logger.Errorf("Re-tries done, couldn't connect to database")
		return nil, errors.New("DB Connection Error")
	}
	s.logger.Infof("Connection to MariaDB Successful")
	s.db = db

	return s, nil
}

func (s *Storage) CreateProd(prod *api.Product) (string, *models.Error) {
	var dbErr models.Error
	var mysqlErr *sqlDrvMySql.MySQLError

	id := storage.GetID()

	prd := product{
		Id:      id,
		Name:    prod.Name,
		Details: prod.Details,
	}

	err := s.db.Create(&prd).Error
	if err != nil {
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			s.logger.Errorf("Key duplication Error: %s while adding new Prod: %v to DB",
				mysqlErr.Error(), prd)

			var value, key string

			fmt.Sscanf(mysqlErr.Error(), "Error 1062 (23000): Duplicate entry %s for key %s", &value, &key)
			if key == "'PRIMARY'" {
				s.logger.Debugf("Re-attempting to add Prod with new Id")

				prd.Id = storage.GetID()
				err := s.db.Create(&prd).Error
				if err != nil {
					dbErr = models.Error{
						Type: models.InternalProcessingError,
					}
					s.logger.Errorf("Error: %s while adding new Prod: %v to DB",
						mysqlErr.Error(), prd)
				} else {
					s.logger.Infof("Successfully Added Prod: %v to DB in 2nd attempt",
						prd)
				}
			} else {
				dbErr = models.Error{
					Type:    models.DuplicateResourceCreationError,
					Message: fmt.Sprintf("Prod already exists, if it is a new Prod, change its details"),
				}
				s.logger.Errorf("Error: %s while adding new Prod: %v to DB",
					mysqlErr.Error(), prd)
			}
		} else {
			dbErr = models.Error{
				Type: models.InternalProcessingError,
			}
			s.logger.Errorf("Error: %s while adding new Prod: %v to DB",
				mysqlErr.Error(), prd)
		}
	} else {
		s.logger.Infof("Successfully Added Prod: %v to DB", prd)
	}

	if err != nil {
		return "", &dbErr
	}

	return prd.Id, nil
}
