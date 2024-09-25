package config

import "time"

type Config struct {
	MaxAttempts int
	Timeout     time.Duration
}
