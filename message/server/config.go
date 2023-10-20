package server

import (
	"time"
)

type Config struct {
	// Connection
	Network  string
	Address  string
	Deadline time.Time
	// Misc
	Log     Log
	Handler Handler
}

func NewConfig(address string) *Config {
	c := &Config{
		Address: address,
	}
	c.setDefaults()
	return c
}

func (c *Config) setDefaults() {
	config := c
	// Config defaults
	if config.Log == nil {
		config.Log = NewLogger()
	}
	if config.Handler == nil {
		config.Handler = NewMux() // handler interface for ServeConn
	}
	if config.Network == "" {
		config.Network = "tcp"
	}
	if config.Address == "" {
		config.Address = "localhost:8080"
	}
	if config.Deadline.IsZero() {
		config.Deadline = time.Now().Add(5 * time.Minute)
	}
}
