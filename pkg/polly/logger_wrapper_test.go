package polly

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestLogger(_ *testing.T) {
	zLogger := zerolog.Logger{}
	lw := LoggerWrapper{
		delegate: &zLogger,
	}
	lw.Printf("Hello %s", "world")
	le := ErrorLoggerWrapper{
		delegate: &zLogger,
	}
	le.Printf("Hello %s", "error")
}
