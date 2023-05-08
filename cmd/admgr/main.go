package main

import (
	"fmt"
	"log"

	"github.com/kiran-anand14/admgr/internal/pkg/core"
	"github.com/kiran-anand14/admgr/internal/pkg/http/rest"
	"github.com/kiran-anand14/admgr/internal/pkg/models"
	"github.com/kiran-anand14/admgr/internal/pkg/storage/mysql"
)

var cnf *Config

func init() {
	cnf = InitializeConfig()
	InitializeLogger(cnf)

	logger.Infof("Initializing admgr Instance: %s", cnf.InstanceId)
}

func main() {
	log.SetFlags(0)

	addr := fmt.Sprintf("%s:%s", cnf.Host, cnf.Port)

	var service core.Service
	var accountService core.AccountingService

	dbConf := models.DBConf{
		Host:     cnf.DB.Host,
		Port:     cnf.DB.Port,
		Name:     cnf.DB.Name,
		Username: cnf.DB.Username,
		Password: cnf.DB.Password,
	}
	s, err := mysql.NewStorage(logger, &dbConf)
	if err != nil {
		logger.Errorf("%s", err.Error())
		return
	}

	accountService = core.NewAccountingService(cnf.Accounting.Host, cnf.Accounting.Port, cnf.InstanceId, logger)
	service = core.NewService(s, accountService, logger)

	r, _ := rest.Handler(logger, service)

	log.Fatal(r.Run(addr))
}
