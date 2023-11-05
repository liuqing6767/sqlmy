package sqlmy

import (
	"context"
	"fmt"
	"log"
)

var logger Logger = &DumbLogger{}

func SetLogger(log Logger) {
	logger = log
}

type LogLevel int

var (
	LogLevelDebug LogLevel = -0
	LogLevelInfo  LogLevel = 0
	LogLevelError LogLevel = 1
)

type Logger interface {
	SetLevel(level LogLevel)
	ValidateLevel(level LogLevel) bool

	Info(ctx context.Context, format string, args ...interface{})
	Error(ctx context.Context, format string, args ...interface{})
}

var (
	_ Logger = &DumbLogger{}
	_ Logger = &StdLogger{}
)

type DumbLogger struct{}

func (dl *DumbLogger) SetLevel(level LogLevel)                                       {}
func (dl *DumbLogger) ValidateLevel(level LogLevel) bool                             { return false }
func (dl *DumbLogger) Info(ctx context.Context, format string, args ...interface{})  {}
func (dl *DumbLogger) Error(ctx context.Context, format string, args ...interface{}) {}

type StdLogger struct {
	Logger   *log.Logger
	logLevel LogLevel
}

func (sl *StdLogger) SetLevel(level LogLevel) { sl.logLevel = level }
func (sl *StdLogger) ValidateLevel(level LogLevel) bool {
	return sl.logLevel >= level
}
func (sl *StdLogger) Info(ctx context.Context, format string, args ...interface{}) {
	sl.output(ctx, LogLevelInfo, format, args...)
}
func (sl *StdLogger) Error(ctx context.Context, format string, args ...interface{}) {
	sl.output(ctx, LogLevelError, format, args...)
}
func (sl *StdLogger) output(ctx context.Context, level LogLevel, format string, args ...interface{}) {
	if !sl.ValidateLevel(level) {
		return
	}

	logPrefix := `[INFO]`
	if logID := GetLogID(ctx); logID != "" {
		logPrefix = "[INFO] [" + logID + "]"
	}
	format = logPrefix + format

	logger := sl.Logger
	if logger == nil {
		logger = log.Default()
	}

	format = format + `\n`
	logger.Output(4, fmt.Sprintf(format, args...))
}
