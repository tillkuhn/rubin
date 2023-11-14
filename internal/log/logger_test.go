package log

import (
	"os"
	"testing"

	"go.uber.org/zap/zapcore"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestProduceMessageOKDebugLevel(t *testing.T) {
	l := NewAtLevel("DEBUG")
	// l.Errorf("hi %s", "till")
	assert.NotNil(t, l)
	assert.Equal(t, zapcore.DebugLevel.String(), l.Level().String()) // value should be debug
}

func TestEnvLogLevel(t *testing.T) {
	_ = os.Setenv("LOG_LEVEL", "warn")
	defer os.Clearenv() // or next test will still use warn level
	l := New()
	assert.NotNil(t, l)
	assert.Equal(t, zapcore.WarnLevel.String(), l.Level().String()) // value should be debug
}

func TestNewDefault(t *testing.T) {
	l := New()
	assert.NotNil(t, l)
	assert.Equal(t, defaultLevel.String(), l.Level().String()) // value should be debug
}

func TestProduceMessageInvalidLevel(t *testing.T) {
	l := NewAtLevel("INVALID")
	assert.NotNil(t, l)
	assert.Equal(t, zap.NewAtomicLevelAt(-0).Level(), l.Level()) // value should be INFO
	assert.Equal(t, defaultLevel.String(), l.Level().String())   // value should be debug
}
