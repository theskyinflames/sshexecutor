package shared

import (
	"time"
)

type (
	SSHExecutor interface {
		Connect(timeout time.Duration) error
		Close() error
		Execute(parameters string) (string, string, error)
	}
)
