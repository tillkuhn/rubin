package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

// NewAtLevel returns a sugared zap Logger configured for the default level
// if the string cannot be parsed, default level INFO is used
// allowed levels: "DEBUG","INFO","WARN","ERROR,"PANIC","DPANIC","FATAL"
func NewAtLevel(levelStr string) *zap.SugaredLogger {
	logLevel := zapcore.InfoLevel
	if levelStr != "" {
		var err error
		logLevel, err = zapcore.ParseLevel(levelStr)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "fallback to info levem, cannot parse loglevel %s: %v\n", levelStr, err)
		}
	}

	// logConf := zap.NewProductionConfig() // with json encoder default info
	logConf := zap.NewDevelopmentConfig() // with console encoder default debug
	logConf.Level = zap.NewAtomicLevelAt(logLevel)

	// Must is a helper that wraps a call to a function returning (*Logger, error)
	// and panics if the error is non-nil
	logger := zap.Must(logConf.Build())

	return logger.Sugar()
}
