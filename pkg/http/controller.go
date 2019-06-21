package http

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/theskyinflames/sshexecutor/pkg/config"
	"github.com/theskyinflames/sshexecutor/pkg/shared"
)

type (
	RecipeExecutor interface {
		Execute(sshExecutor shared.SSHExecutor, recipe []string, timeout time.Duration) ([]string, string, error)
	}

	Controller struct {
		receiptExecutor RecipeExecutor

		cfg *config.Config
		log *logrus.Logger
	}
)

func NewController(receiptExecutor RecipeExecutor, log *logrus.Logger) *Controller {
	return &Controller{receiptExecutor: receiptExecutor, log: log}
}

func (c *Controller) RunRecipe(sshExecutor shared.SSHExecutor, runRecipe *SSHRecipeRq) (*SSHRecipeRs, error) {

	rs, rserr, err := c.receiptExecutor.Execute(sshExecutor, runRecipe.Recipe, runRecipe.Timeout*time.Second)

	return &SSHRecipeRs{
		Response:    rs,
		ResponseErr: rserr,
		Error:       getErrorḾsg(err),
	}, err
}

func getErrorḾsg(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
