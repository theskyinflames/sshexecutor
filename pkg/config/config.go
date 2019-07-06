package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/codingconcepts/env"
)

type (
	Config struct {
		APIHost                  string        `env:"SSH_EXECUTOR_API_HOST" required:"true"`
		APIPort                  string        `env:"SSH_EXECUTOR_API_PORT" required:"true"`
		DefaultSSHCommandTimeout time.Duration `env:"SSH_EXECUTOR_API_DEFAULT_SSH_TIMEOUT" required:"true"`
		SSHUser                  string        `env:"SSH_EXECUTOR_USER" required:"true"`
		SSHPassword              string        `env:"SSH_EXECUTOR_PASSWORD" required:"false"`
		SSHPrivateKeyPath        string        `env:"SSH_EXECUTOR_PRIVATE_KEY_PATH" required:"false"`
		SSHPrivateKey            []byte
	}
)

func (c *Config) Load() error {
	return env.Set(c)
}

func (c *Config) LoadSSHPublicKey() (err error) {
	c.SSHPrivateKey, err = ioutil.ReadFile(c.SSHPrivateKeyPath)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) Validate() error {
	if len(c.SSHPassword) == 0 && len(c.SSHPrivateKeyPath) == 0 {
		return errors.New("either ssh password or public key file path must be provided")
	}
	if len(c.SSHPrivateKeyPath) > 0 && len(c.SSHPrivateKey) == 0 {
		return fmt.Errorf("public key has not been loaded from %s", c.SSHPrivateKeyPath)
	}
	return nil
}
