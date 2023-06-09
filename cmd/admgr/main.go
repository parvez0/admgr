package main

import (
	"fmt"
	"github.com/kiran-anand14/admgr/internal/pkg/accounting"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/kiran-anand14/admgr/internal/pkg/core"
	"github.com/kiran-anand14/admgr/internal/pkg/http/rest"
	"github.com/kiran-anand14/admgr/internal/pkg/models"
	"github.com/kiran-anand14/admgr/internal/pkg/storage/mysql"
)

var cnf *Config

func main() {

	cnf = InitializeConfig()

	absPath := ""
	if config.Logger.OutputFilePath != "" {
		var err error
		absPath, err = filepath.Abs(config.Logger.OutputFilePath)
		if err != nil {
			panic(fmt.Errorf("failed to load logfile : %s", err.Error()))
		}
		path := strings.Split(absPath, "/")
		_, err = os.Stat(strings.Join(path[:len(path)-1], "/"))
		if err != nil {
			panic(fmt.Errorf("failed to load logfile : %s", err.Error()))
		}
	}
	fmt.Println("Output log filepath: ", absPath)
	fd, err := os.OpenFile(absPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("failed to create log file %s", err.Error()))
	}
	defer fd.Close()
	writer := io.MultiWriter(os.Stdout, fd)
	InitializeLogger(cnf, writer)

	logger.Infof("Initializing admgr Instance: %s", cnf.InstanceId)

	log.SetFlags(0)

	addr := fmt.Sprintf("%s:%s", cnf.Host, cnf.Port)

	var service core.Service
	var accountService accounting.AccountingService

	dbConf := models.DBConf(cnf.DB)
	s, err := mysql.NewStorage(logger, writer, cnf.Logger.Level, &dbConf)
	if err != nil {
		logger.Errorf("%s", err.Error())
		return
	}
	acntServiceConf := models.AccountingServiceConf(cnf.Accounting)
	accountService = accounting.NewAccountingService(logger, acntServiceConf, cnf.InstanceId)
	service = core.NewService(s, accountService, logger)

	r, _ := rest.Handler(logger, service, writer)

	log.Fatal(r.Run(addr))
}
