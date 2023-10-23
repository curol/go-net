package server

import (
	"fmt"
	"log"
	"time"
)

type Log interface {
	Status(path, method, remoteAddress string) // Log status
	Fatal(error)                               // Log error and exit
	// TODO: Add more logging methods
}

type logger struct{}

// NewLogger returns a new logger
func NewLogger() *logger {
	return &logger{}
}

// Log logs connection status
func (l *logger) Status(path, method, remoteAddress string) {
	// Time
	now := time.Now()
	timeFormat := now.Format("2006-01-02 15:04:05")
	// Connection
	// Request
	s := "%s Status: %s (path: %s) (method: %s)\n"
	fmt.Printf(s, timeFormat, remoteAddress, path, method)
}

// Fatal logs error and exits
func (l *logger) Fatal(err error) {
	log.Fatal(err)
}
