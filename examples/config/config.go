package config

import "time"

// Service is service config.
type Service struct {
	Name    string
	Version string
}

// Server is server config.
type Server struct {
	Network string
	Address string
	Timeout time.Duration
}
