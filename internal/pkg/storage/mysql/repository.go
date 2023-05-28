package mysql

import (
	"errors"
	"fmt"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"io"
	"log"
	"net"
	"strings"
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
	loglevel string
}

func NewStorage(_log *logrus.Logger, writer io.Writer, logLevel string, dbConf *models.DBConf) (*Storage, error) {
	s := new(Storage)

	s.logger = _log

	dsn := dbConf.Username + ":" + dbConf.Password + "@tcp" + "(" + dbConf.Host +
		":" + dbConf.Port + ")/" + dbConf.Name + "?" + "charset=utf8mb4&parseTime=True&loc=Local&clientFoundRows=true&timeout=60s"

	s.logger.Debugf("Database Connection String: %s", dsn)

	var db *gorm.DB
	var err error
	var retryCount uint8

	gormLogger := logger.New(
		log.New(writer, "", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200,
			LogLevel:                  getLogLevel(logLevel),
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)
	for {
		retryCount++
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: gormLogger,
		})
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
	if strings.ToLower(logLevel) == "debug" {
		db = db.Debug()
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

func getLogLevel(lvl string) logger.LogLevel {
	switch strings.ToLower(lvl) {
	case "info":
		return logger.Info
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	}
	return logger.Info
}

func (s *Storage) Create(records interface{}) (int, error) {
	var dbErr error
	var mysqlErr *sqlDrvMySql.MySQLError

	res := s.db.Create(records)
	if res.Error != nil {
		err := res.Error
		switch err.(type) {
		case *sqlDrvMySql.MySQLError:
			errors.As(err, &mysqlErr)
			if mysqlErr.Number == 1062 {
				s.logger.Errorf("DbInsertFailed:: key duplication Error: %s while "+
					"adding new record: %+v", mysqlErr.Error(), records)
				dbErr = models.NewError("FailedToCreate:: Duplicate records provided", models.DuplicateResourceCreationError)
				break
			}
			s.logger.Errorf("DbInsertFailed:: [Code: %d, Error: %s]", mysqlErr.Number, mysqlErr.Message)
			dbErr = models.NewError("FailedToCreate:: Internal server error", models.InternalProcessingError)
		default:
			s.logger.Errorf("DbInsertFailed:: [Code: %d, Error: %s]", -1, err)
			dbErr = models.NewError("FailedToCreate:: Internal server error", models.InternalProcessingError)
		}
		return 0, dbErr
	}
	s.logger.Infof("Create:: Total %d records created successfully", res.RowsAffected)
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
			s.logger.Errorf("UpdateRecordsFailed:: %s :: %+v", res.Error, slot)
			dbError = models.NewError("PatchFailed:: Internal server error", models.InternalProcessingError)
			break
		}
		if res.RowsAffected == 0 {
			s.logger.Debug("UpdateRecords:: Slot not found in DB [", slot.ToString(), "] reverting changes")
			dbError = models.NewError(fmt.Sprintf("Slot details not found %s", slot.ToString()), models.ActionForbidden)
			break
		}
		affectedRows += int(res.RowsAffected)
	}
	if dbError != nil {
		tx.Rollback()
		return 0, dbError
	}
	s.logger.Infof("Update:: Total %d records affected, updated successfully", affectedRows)
	return affectedRows, tx.Commit().Error
}

func (s *Storage) SearchSlotsInRange(options *GetOptions) ([]*Slot, error) {
	var slots []*Slot
	query := s.db.Model(&Slot{}).
		Where("date BETWEEN ? AND ?", options.StartDate.Format(time.DateOnly), options.EndDate.Format(time.DateOnly))
	if options.PositionStart != "" && options.PositionEnd != "" {
		query = query.Where("position BETWEEN ? AND ?", options.PositionStart, options.PositionEnd)
	}
	if options.Status != "" {
		query = query.Where("status = ?", options.Status)
	}
	if options.Uid != "" {
		query = query.Where("booked_by = ?", options.Uid)
	}
	res := query.Find(&slots)
	if res.Error != nil {
		s.logger.Errorf("SearchSlotsInRange::[%+v]", options)
		return nil, models.NewError(
			"GetSlotsFailed:: Internal server error",
			models.InternalProcessingError,
		)
	}
	s.logger.Infof("SearchSlotsInRange:: Total %d records found", res.RowsAffected)
	return slots, nil
}

func (s *Storage) SearchSlotsByStatus(options *GetOptions) ([]*Slot, error) {
	var slots []*Slot
	db := s.db.Model(&Slot{}).Where("status = ?", options.Status)
	if options.PreloadTransaction {
		db = db.Preload("Transaction")
	}
	if err := db.Find(&slots).Error; err != nil {
		return nil, models.NewError(
			fmt.Sprintf("SearchSlotsByStatus: [Status: %s, Error: %s]", options.Status, err),
			models.InternalProcessingError,
		)
	}
	return slots, nil
}

func (s *Storage) UpdateSlotsStatus(slots []*Slot, lastStatus, newStatus string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		for i, slot := range slots {
			var resSlot Slot
			if err := tx.Model(&Slot{}).
				Where("date = ? AND position = ? AND status = ?", slot.Date.Format(time.DateOnly), slot.Position, lastStatus).
				First(&resSlot).
				Error; err != nil {
				return models.NewError(
					fmt.Sprintf("SlotNotFound:: Slot cannot be booked [date: %s, position: %v]", models.DateToString(*slot.Date), *slot.Position),
					models.ActionForbidden,
				)
			}
			resSlot.Status = &newStatus
			if err := s.db.Save(&resSlot).Error; err != nil {
				s.logger.Errorf("SlotUpdateFailed:: [Error: %s, Slot: %+v]", err.Error(), resSlot)
				return models.NewError(
					fmt.Sprintf("SlotUpdateFailed:: Internal server error"),
					models.InternalProcessingError,
				)
			}
			slots[i] = &resSlot
		}
		return nil
	})
}

func (s *Storage) Delete(records interface{}) (int, error) {
	res := s.db.Delete(records)
	if res.Error != nil {
		s.logger.Errorf("DeleteRecordsFailed:: [Error: %s, Records: %+v]", res.Error, records)
		return 0, models.NewError("DeleteFailed:: Internal server error", models.InternalProcessingError)
	}
	s.logger.Infof("Delete:: Total %d matching records deleted", res.RowsAffected)
	return int(res.RowsAffected), nil
}

func (s *Storage) DropAll() error {
	return s.db.Migrator().DropTable(&Transaction{}, &Slot{})
}

func (s *Storage) Initialize() error {
	return s.db.AutoMigrate(&Transaction{}, &Slot{})
}
