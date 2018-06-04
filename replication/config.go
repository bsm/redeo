package replication

import (
	"log"
	"os"
	"time"
)

// Logger interface
type Logger interface {
	Printf(format string, v ...interface{})
}

// Config holds the replication configuration
type Config struct {
	// Size of the replication backlog. Must be at least 128KiB.
	// Default: 2MiB
	BacklogSize int

	// Timeout for master/slave establishing connection.
	// Default: 10s
	DialTimeout time.Duration

	// Logger is a custom logger
	// Default: log.New(os.Stderr, "[redeo]", log.LstdFlags)
	Logger Logger
}

func (c *Config) norm() {
	if c.BacklogSize < minBacklogSize {
		c.BacklogSize = 2 * 1024 * 1024
	}
	if c.DialTimeout < 1 {
		c.DialTimeout = 10 * time.Second
	}
	if c.Logger == nil {
		c.Logger = log.New(os.Stderr, "[redeo]", log.LstdFlags)
	}
}
