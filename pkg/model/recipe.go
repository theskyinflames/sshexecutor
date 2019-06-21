package model

import (
	"github.com/sirupsen/logrus"

	"github.com/theskyinflames/sshexecutor/pkg/shared"
)

type (
	Recipe struct {
		Recipe []string

		log *logrus.Logger
	}
)

func (r Recipe) Execute(server shared.SSHExecutor) (stdOut []string, stdErr string, err error) {
	stdOut = make([]string, len(r.Recipe))
	for z, command := range r.Recipe {
		stdOut[z], stdErr, err = server.Execute(command)
		if err != nil {
			return nil, "", err
		}
	}
	return stdOut, stdErr, nil
}
