package log

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const defaultLevel = zapcore.InfoLevel

// New delegated to NewAzLevel and returns a sugared zap Logger configured for the default level
// which is either determined by the environment variable LOG_LEVEL, or defaults to 'info'
func New() *zap.SugaredLogger {
	ll := os.Getenv("LOG_LEVEL")
	if ll != "" {
		return NewAtLevel(ll)
	}
	return NewAtLevel(defaultLevel.String())
}

// NewAtLevel returns a sugared zap Logger configured for the given level
// if the string is empty or cannot be parsed, default level INFO is used
// allowed levels: "DEBUG","INFO","WARN","ERROR,"PANIC","DPANIC","FATAL"
func NewAtLevel(levelStr string) *zap.SugaredLogger {
	logLevel := defaultLevel
	if levelStr != "" {
		var err error
		logLevel, err = zapcore.ParseLevel(levelStr)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "cannot parse loglevel '%s', callback to level 'info': %v\n", levelStr, err)
		}
	}

	// Use logger preset NewDevelopmentConfig or NewProductionConfig() (with json encoder default info)
	// see https://blog.sandipb.net/2018/05/02/using-zap-simple-use-cases/
	// Advanced: https://github.com/sandipb/zap-examples/tree/master/src/customlogger#using-the-zap-config-struct-to-create-a-logger
	// The Development logger:
	//   Prints a stack trace from Warn level and up.
	//   Always prints the package/file/line number (the caller)
	//   Tacks any extra fields as a json string at the end of the line
	//   Prints the level names in uppercase
	//   Prints timestamp in ISO8601 format with milliseconds
	//
	// logConf := zap.NewDevelopmentConfig() // with console encoder default debug
	logConf := zap.Config{
		Level: zap.NewAtomicLevelAt(logLevel),
		// Development:      true,
		// DisableCaller stops annotating logs with the calling function's file, line no etc.
		DisableCaller: true,
		// DisableStacktrace completely disables automatic stacktrace capturing. default: warn, error
		DisableStacktrace: true,
		Encoding:          "console", // or "json"
		EncoderConfig:     zap.NewDevelopmentEncoderConfig(),
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
	}

	// Configure date format https://github.com/uber-go/zap/issues/485#issuecomment-834021392
	// time.RFC3339 or time.RubyDate or "2006-01-02 15:04:05" or even freaking time.Kitchen
	logConf.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.Kitchen)

	// Configure Color, see https://github.com/uber-go/zap/issues/648#issuecomment-492481968
	logConf.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	// logConf.Level = zap.NewAtomicLevelAt(logLevel)

	// DisableStacktrace completely disables automatic stacktrace capturing. By
	// default, stacktraces are captured for WarnLevel and above logs in
	// development and ErrorLevel and above in production.
	// logConf.DisableStacktrace = true

	// Must is a helper that wraps a call to a function returning (*Logger, error)
	// and panics if the error is non-nil
	logger := zap.Must(logConf.Build())

	return logger.Sugar()
}
