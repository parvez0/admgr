package mysql

import (
	"errors"
	"fmt"
	"gorm.io/gorm/clause"
	"net"
	"time"

	"github.com/sirupsen/logrus"

	sqlDrvMySql "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/kiran-anand14/admgr/internal/pkg/models"
)

type Storage struct {
	logger   *logrus.Logger
	seedFile string
	db       *gorm.DB
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
			if _, ok := err.(*net.OpError); ok {
				return nil, errors.New(fmt.Sprintf("DBConnectionFailed::[Config: %+v, Error: %s]", dbConf, err))
			}
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
	s.logger.Infof("Connection to MariaDB Successfull, initiating db seeding")
	err = db.AutoMigrate(&Slot{}, &Transaction{})
	// Add foreign key constraint
	if err != nil {
		return nil, errors.New(fmt.Sprintf("DBSeeding failed with error: %s", err))
	}
	s.logger.Infof("DB Seeding succeded")
	s.db = db
	return s, nil
}

func (s *Storage) Create(records interface{}) (int, error) {
	var dbErr error
	var mysqlErr *sqlDrvMySql.MySQLError

	res := s.db.Create(records)
	if res.Error != nil {
		err := res.Error
		errMsg := "CreateInsertFailed::"
		switch err.(type) {
		case *sqlDrvMySql.MySQLError:
			errors.As(err, &mysqlErr)
			if mysqlErr.Number == 1062 {
				errMsg = fmt.Sprintf("%s key duplication Error: %s while "+
					"adding new record: %v", errMsg, mysqlErr.Error(), records)
				dbErr = models.NewError(errMsg, models.DuplicateResourceCreationError)
				break
			}
			errMsg = fmt.Sprintf("[Code: %d, Error: %s]", mysqlErr.Number, mysqlErr.Message)
			dbErr = models.NewError(errMsg, models.InternalProcessingError)
		default:
			errMsg = fmt.Sprintf("%s %s", errMsg, err.Error())
			dbErr = models.NewError(errMsg, models.InternalProcessingError)
		}
		return 0, dbErr
	}
	s.logger.Infof("Total %d records created successfully", res.RowsAffected)
	return int(res.RowsAffected), nil
}

func (s *Storage) UpdateSlots(slots []*Slot) (int, error) {
	var dbError error
	tx, affectedRows := s.db.Begin(), 0
	for _, slot := range slots {
		res := tx.Model(&Slot{}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("date = ? and position = ?", slot.Date.Format(time.DateOnly), slot.Position).
			Omit("date", "position").
			Updates(slot)
		if res.Error != nil {
			dbError = models.NewError(fmt.Sprintf("UpdateRecordsFailed:: %s :: %+v", res.Error, slot), models.InternalProcessingError)
			break
		}
		if res.RowsAffected == 0 {
			dbError = models.NewError(fmt.Sprintf("UpdateRecordsFailed::RecordNotFound::%+v", slot), models.ActionForbidden)
			break
		}
		affectedRows += int(res.RowsAffected)
	}
	if dbError != nil {
		tx.Rollback()
		return 0, dbError
	}
	s.logger.Infof("Total %d records affected, updated successfully", affectedRows)
	return affectedRows, tx.Commit().Error
}

func (s *Storage) SearchSlots(filters map[string]interface{}) ([]*Slot, error) {
	var slots []*Slot
	query := s.db.Model(&Slot{})
	for key, value := range filters {
		query = query.Debug().Where(fmt.Sprintf("%s = ?", key), value)
	}
	res := query.Find(&slots)
	if res.Error != nil {
		return nil, models.NewError(
			fmt.Sprintf("SearchRecordFailed::%s", res.Error),
			models.InternalProcessingError,
		)
	}
	return slots, nil
}

func (s *Storage) SearchSlotsByPrimaryKeyAndStatus(date time.Time, position int32, status string) (*Slot, error) {
	var slot Slot
	if err := s.db.Model(&Slot{}).
		Debug().
		Where("date = ? AND position = ? AND status = ?", date.Format(time.DateOnly), position, status).
		First(&slot).Error; err != nil {
		return nil, models.NewError(
			fmt.Sprintf("SearchFirstRecordFailed:: %s", err.Error()),
			models.InternalProcessingError,
		)
	}
	return &slot, nil
}

func (s *Storage) UpdateSlotsStatus(slots []*Slot, lastStatus, newStatus string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		for i, slot := range slots {
			var resSlot Slot
			if err := tx.Model(&Slot{}).
				Debug().
				Where("date = ? AND position = ? AND status = ?", slot.Date.Format(time.DateOnly), slot.Position, lastStatus).
				First(&resSlot).
				Error; err != nil {
				return models.NewError(
					fmt.Sprintf("SearchFailed:: Record cannot be booked %+v", slot),
					models.ActionForbidden,
				)
			}
			resSlot.Status = &newStatus
			if err := s.db.Save(&resSlot).Error; err != nil {
				return models.NewError(
					fmt.Sprintf("UpdateStatusFailed:: [Error: %s, Slot: %+v]", err.Error(), resSlot),
					models.InternalProcessingError,
				)
			}
			slots[i] = &resSlot
		}
		return nil
	})
}

func (s *Storage) DropAll() error {
	return s.db.Migrator().DropTable(&Transaction{}, &Slot{})
}
