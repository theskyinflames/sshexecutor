package http

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/expvar"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/theskyinflames/sshexecutor/pkg/config"
	"github.com/theskyinflames/sshexecutor/pkg/model"
	"github.com/theskyinflames/sshexecutor/pkg/shared"
	"github.com/theskyinflames/sshexecutor/pkg/ssh"
)

type (
	RecipeRunController interface {
		RunRecipe(shared.SSHExecutor, *SSHRecipeRq) (*SSHRecipeRs, error)
	}

	SSHRecipeRq struct {
		Host    string        `json:"host" binding:"required"`
		Port    int           `json:"port" binding:"required"`
		Recipe  []string      `json:"recipe" binding:"required"`
		Timeout time.Duration `json:"timeout"`
	}

	SSHRecipeRs struct {
		Response    []string `json:"response"`
		ResponseErr string   `json:"responseErr"`
		Error       string   `json:"error"`
	}

	API struct {
		controller RecipeRunController

		cfg *config.Config
		log *logrus.Logger
	}
)

func (r SSHRecipeRq) Validate() error {
	if len(r.Recipe) == 0 {
		return errors.New("at least, a ssh command must be provided")
	}
	return nil
}

func NewAPI(controller RecipeRunController, cfg *config.Config, log *logrus.Logger) *API {
	return &API{controller: controller, cfg: cfg, log: log}
}

func (api *API) Start() error {

	router := gin.New()

	router.Use(gin.Recovery())

	// Add a ginrus middleware, which:
	//   - Logs all requests, like a combined access and error log.
	//   - Logs to stdout.
	//   - RFC3339 with UTC time format.
	router.Use(ginrus.Ginrus(api.log, time.RFC3339, true))

	router.GET("/debug/vars", expvar.Handler())
	router.GET("/check", expvar.Handler())
	router.POST("/runreceipt", api.runRecipe)

	router.Run(fmt.Sprintf("%s:%s", api.cfg.APIHost, api.cfg.APIPort))

	return nil
}

func (api *API) runRecipe(c *gin.Context) {

	var rq SSHRecipeRq
	if err := c.ShouldBindJSON(&rq); err != nil {
		api.log.WithFields(logrus.Fields{"error": err.Error()}).Error("request not valid")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := rq.Validate()
	if err != nil {
		api.log.WithFields(logrus.Fields{"error": err.Error()}).Error("request not valid")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ssh := ssh.NewSSHExecutorServer(rq.Host, rq.Port, api.cfg.SSHUser, api.cfg.SSHPassword, api.cfg.SSHPublicKey, api.cfg, api.log)
	result, err := api.controller.RunRecipe(ssh, &rq)
	if err != nil {
		api.log.WithFields(logrus.Fields{"error": err.Error()}).Error("something went wrong starting the running the recipe")

		switch err {
		case model.ErrRequestTimeout:
			c.JSON(http.StatusRequestTimeout, gin.H{"status": http.StatusRequestTimeout, "error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "error": err.Error()})
		}
	} else {
		c.JSON(http.StatusOK, result)
	}
}
