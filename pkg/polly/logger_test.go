package polly

import (
	"testing"

	"github.com/tillkuhn/rubin/internal/log"
)

func TestLogger(_ *testing.T) {
	zLogger := log.NewAtLevel("")
	lw := LoggerWrapper{
		delegate: zLogger,
	}
	lw.Printf("Hello %s", "world")
}
