// Copyright (C) 2025, ADXYZ Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package log

import (
	"github.com/luxfi/node/utils/logging"
	"go.uber.org/zap"
)

// Logger is a wrapper around luxfi's Logger interface
type Logger interface {
	Debug(msg string)
	Info(msg string) 
	Warn(msg string)
	Error(msg string)
	Fatal(msg string)
	Sync() error
}

// luxLogger wraps luxfi/node's Logger
type luxLogger struct {
	log logging.Logger
}

// New creates a new logger using luxfi's logging
func New() Logger {
	return NewWithLevel("info")
}

// NewWithLevel creates a new logger with specific level
func NewWithLevel(level string) Logger {
	// Map string level to luxfi's Level type
	lvl := logging.Info
	switch level {
	case "debug":
		lvl = logging.Debug
	case "info":
		lvl = logging.Info
	case "warn":
		lvl = logging.Warn
	case "error":
		lvl = logging.Error
	case "fatal":
		lvl = logging.Fatal
	}

	// Create logger with default config
	config := logging.Config{
		DisplayLevel: lvl,
		LogLevel:     lvl,
		DisableWriterDisplaying: false,
	}
	
	factory := logging.NewFactory(config)
	log, err := factory.Make("adx")
	if err != nil {
		return &noOpLogger{}
	}
	
	return &luxLogger{log: log}
}

// NoOp returns a no-op logger
func NoOp() Logger {
	return &noOpLogger{}
}

// NoLog is a no-op logger instance
var NoLog = NoOp()

// NewLogger creates a new logger with a name
func NewLogger(name string) Logger {
	config := logging.Config{
		DisplayLevel: logging.Info,
		LogLevel:     logging.Info,
	}
	
	factory := logging.NewFactory(config)
	log, err := factory.Make(name)
	if err != nil {
		return &noOpLogger{}
	}
	
	return &luxLogger{log: log}
}

// Debug logs a debug message
func (l *luxLogger) Debug(msg string) {
	l.log.Debug(msg)
}

// Info logs an info message
func (l *luxLogger) Info(msg string) {
	l.log.Info(msg)
}

// Warn logs a warning message  
func (l *luxLogger) Warn(msg string) {
	l.log.Warn(msg)
}

// Error logs an error message
func (l *luxLogger) Error(msg string) {
	l.log.Error(msg)
}

// Fatal logs a fatal message and exits
func (l *luxLogger) Fatal(msg string) {
	l.log.Fatal(msg)
}

// Sync flushes any buffered log entries
func (l *luxLogger) Sync() error {
	l.log.Stop()
	return nil
}

// noOpLogger is a logger that does nothing
type noOpLogger struct{}

func (n *noOpLogger) Debug(msg string) {}
func (n *noOpLogger) Info(msg string)  {}
func (n *noOpLogger) Warn(msg string)  {}
func (n *noOpLogger) Error(msg string) {}
func (n *noOpLogger) Fatal(msg string) {}
func (n *noOpLogger) Sync() error      { return nil }

// For compatibility with zap.Field usage in some places
func String(key, val string) zap.Field {
	return zap.String(key, val)
}

func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

func Error(err error) zap.Field {
	return zap.Error(err)
}