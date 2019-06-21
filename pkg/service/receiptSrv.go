package service

import (
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/theskyinflames/sshexecutor/pkg/model"
	"github.com/theskyinflames/sshexecutor/pkg/shared"
)

type (
	ExecutorSrv struct {
		log *logrus.Logger
	}
)

func NewExecutorSrv(log *logrus.Logger) *ExecutorSrv {
	return &ExecutorSrv{log: log}
}

func (re *ExecutorSrv) Execute(sshExecutor shared.SSHExecutor, rqreceipt []string, timeout time.Duration) (receiptRs []string, receiptErr string, err error) {

	err = sshExecutor.Connect(timeout)
	if err != nil {
		return nil, "", errors.Wrap(err, "some wen wrong when trying to connect to target server")
	}
	defer sshExecutor.Close()

	recipe := &model.Recipe{Recipe: rqreceipt}
	receiptRs, receiptErr, err = recipe.Execute(sshExecutor)
	if err != nil {
		return nil, "", errors.Wrap(err, "some wen wrong when trying to execute the recipe")
	}

	return receiptRs, receiptErr, nil
}
