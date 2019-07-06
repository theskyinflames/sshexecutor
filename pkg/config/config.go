package config

import (
	"time"

	"github.com/codingconcepts/env"
)

type (
	Config struct {
		APIHost                  string        `env:"SSH_EXECUTOR_API_HOST" required:"true"`
		APIPort                  string        `env:"SSH_EXECUTOR_API_PORT" required:"true"`
		DefaultSSHCommandTimeout time.Duration `env:"SSH_EXECUTOR_API_DEFAULT_SSH_TIMEOUT" required:"true"`
		SSHUser                  string        `env:"SSH_EXECUTOR_USER" required:"true"`
		SSHPassword              string        `env:"SSH_EXECUTOR_PASSWORD" required:"true"`
	}
)

func (c *Config) Load() error {
	return env.Set(c)
}
