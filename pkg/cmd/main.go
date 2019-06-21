package main

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/theskyinflames/sshexecutor/pkg/config"
	"github.com/theskyinflames/sshexecutor/pkg/http"
	"github.com/theskyinflames/sshexecutor/pkg/service"
)

type (
	API interface {
		Start() error
	}

	UTCFormatter struct {
		logrus.Formatter
	}
)

func (u UTCFormatter) Format(e *logrus.Entry) ([]byte, error) {
	e.Time = e.Time.UTC()
	return u.Formatter.Format(e)
}

func getLogger() *logrus.Logger {
	// Set logging service
	log := logrus.New()
	log.SetFormatter(UTCFormatter{
		&logrus.JSONFormatter{
			PrettyPrint: false,
		},
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)
	return log
}

func getConfig() (*config.Config, error) {
	cfg := &config.Config{}
	err := cfg.Load()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func main() {

	cfg, err := getConfig()
	if err != nil {
		panic(err)
	}

	log := getLogger()

	executorSrv := service.NewExecutorSrv(log)

	controller := http.NewController(executorSrv, log)

	var api API = http.NewAPI(controller, cfg, log)

	log.WithFields(logrus.Fields{"version": "0.0.0-1"}).Info("started")
	err = api.Start()
	if err != nil {
		log.WithFields(logrus.Fields{"err": err.Error()}).Error("something went wrong starting the api")
	}
}
