package polly

import (
	"strings"

	"github.com/rs/zerolog"
)

// LoggerWrapper wraps zerolog logger so we can used it as logger in kafka-go ReaderConfig
// Example:
//
//	r := kafka.NewReader(kafka.ReaderConfig{
//		Logger:      LoggerWrapper{delegate: k.logger},
//	})
type LoggerWrapper struct {
	delegate *zerolog.Logger
}

func (l LoggerWrapper) Printf(format string, v ...interface{}) {
	l.delegate.Printf(strings.TrimSpace("kafka-go: "+format), v...) // trimspace also removed \n
}

type ErrorLoggerWrapper struct {
	delegate *zerolog.Logger
}

func (l ErrorLoggerWrapper) Printf(format string, v ...interface{}) {
	l.delegate.Error().Msgf("kafka-go: "+format, v...)
}
