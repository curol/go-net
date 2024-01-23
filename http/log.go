package http

import (
	"fmt"
	"time"
)

type Log interface {
	Status(path, method, remoteAddress string) // Log status
	Fatal(error)                               // Log error and exit
	// TODO: Add more logging methods
}

type logger struct {
	format string
}

// NewLogger returns a new logger
func NewLogger() *logger {
	return &logger{format: "2006-01-02 15:04:05"}
}

// Log logs connection status
func (l *logger) Status(addr, method, path string) {
	// Color
	green := "\033[32m"
	reset := "\033[0m"
	// Time
	now := time.Now()
	timeFormat := now.Format(l.format)

	s := "%s Status: (remote: %s) (method: %s) (path: %s)\n"
	s = fmt.Sprintf(s, timeFormat, addr, path, method)
	fmt.Println(green + s + reset)
}

// Fatal logs error and exits
func (l *logger) Fatal(err error) {
	// Color
	red := "\033[31m"
	reset := "\033[0m"
	// Time
	now := time.Now()
	timeFormat := now.Format("2006-01-02 15:04:05")

	s := "%s Error: %s %s\n"
	s = fmt.Sprintf(s, timeFormat, err.Error())
	fmt.Printf(red + s + reset)
}
