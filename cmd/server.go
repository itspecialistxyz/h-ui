package cmd

import (
	"errors"
	"fmt"
	"h-ui/dao"
	"h-ui/middleware"
	"h-ui/model/constant"
	"h-ui/router"
	"h-ui/service"
	"h-ui/util"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func runServer(port string) error {
	defer releaseResource()

	// Check for initial setup
	allowedDomainEnv := os.Getenv("HUI_ALLOWED_DOMAIN")
	securityPathEnv := os.Getenv("HUI_SECURITY_PATH")

	if allowedDomainEnv == "" || securityPathEnv == "" {
		// Env vars not set, try to get from DB
		if err := dao.InitSqliteDB(); err != nil {
			logrus.Fatalf("Failed to initialize database for setup check: %v. Please run './h-ui setup'", err)
		}
		allowedDomainDB, _ := dao.GetConfig("key = ?", constant.HUIAllowedDomain)
		securityPathDB, _ := dao.GetConfig("key = ?", constant.HUISecurityPath)
		dao.CloseSqliteDB() // Close DB after checking

		if allowedDomainDB.Value == "" || securityPathDB.Value == "" {
			// Not found in DB either
			fmt.Println("Initial setup required. Please run: ./h-ui setup")
			os.Exit(1)
		}
		// If found in DB, proceed (middleware will pick them up)
		logrus.Info("Domain and security path settings found in database.")
	} else {
		logrus.Info("Domain and security path settings found in environment variables.")
	}
	// At this point, settings are either in env vars or DB, or the program has exited.

	middleware.InitLog()
	service.InitForward()
	if err := initFile(); err != nil {
		return err
	}
	if err := dao.InitSql(port); err != nil {
		return err
	}
	if err := middleware.InitCron(); err != nil {
		return err
	}
	if err := service.InitHysteria2(); err != nil {
		return err
	}
	if err := service.InitTableAndChain(); err != nil {
		logrus.Errorf(err.Error())
	}
	if err := service.InitPortHopping(); err != nil {
		logrus.Errorf(err.Error())
	}
	if err := service.InitTelegramBot(); err != nil {
		logrus.Errorf(err.Error())
	}

	config, err := dao.GetConfig("key = ?", constant.HUIWebContext)
	if err != nil {
		return err
	}

	r := gin.Default()
	router.Router(r, config.Value)

	serverPort, crtPath, keyPath, err := service.GetServerPortAndCert()
	if err != nil {
		return err
	}

	service.InitServer(fmt.Sprintf(":%d", serverPort), r)
	if err := service.StartServer(crtPath, keyPath); err != nil && err != http.ErrServerClosed {
		logrus.Errorf("start server err: %v", err)
		return errors.New("start server err")
	}
	return nil
}

func releaseResource() {
	if err := dao.CloseSqliteDB(); err != nil {
		logrus.Errorf(err.Error())
	}
	if err := service.ReleaseHysteria2(); err != nil {
		logrus.Errorf(err.Error())
	}
	if err := service.RemoveByComment(); err != nil {
		logrus.Errorf(err.Error())
	}
}

func initFile() error {
	var dirs = []string{constant.LogDir, constant.SqliteDBDir, constant.BinDir, constant.ExportPathDir}
	for _, item := range dirs {
		if !util.Exists(item) {
			if err := os.Mkdir(item, os.ModePerm); err != nil {
				logrus.Errorf("%s create err: %v", item, err)
				return errors.New(fmt.Sprintf("%s create err", item))
			}
		}
	}
	return nil
}
