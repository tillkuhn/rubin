package log

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	// Configure date format https://github.com/uber-go/zap/issues/485#issuecomment-834021392
	// time.RFC3339 or time.RubyDate or "2006-01-02 15:04:05" or even freaking time.Kitchen
	logConf.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.Kitchen)
	logConf.Level = zap.NewAtomicLevelAt(logLevel)

	// Must is a helper that wraps a call to a function returning (*Logger, error)
	// and panics if the error is non-nil
	logger := zap.Must(logConf.Build())

	return logger.Sugar()
}
