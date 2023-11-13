package polly

import (
	"go.uber.org/zap"
)

// LoggerWrapper wraps zerolog logger so we can used it as logger in kafka-go ReaderConfig
// Example:
//
//	r := kafka.NewReader(kafka.ReaderConfig{
//		Logger:      LoggerWrapper{delegate: k.logger},
//	})
type LoggerWrapper struct {
	delegate *zap.SugaredLogger
}

func (l LoggerWrapper) Printf(format string, v ...interface{}) {
	l.delegate.Debugf("kafka-go "+format, v...)
}

type ErrorLoggerWrapper struct {
	delegate *zap.SugaredLogger
}

func (l ErrorLoggerWrapper) Printf(format string, v ...interface{}) {
	l.delegate.Errorf("kafka-go "+format, v...)
}
