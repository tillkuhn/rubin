package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestProduceMessageOK(t *testing.T) {
	l := NewAtLevel("DEBUG")
	assert.NotNil(t, l)
	assert.Equal(t, zap.NewAtomicLevelAt(-1).Level(), l.Level())
	l = NewAtLevel("INVALID")
	assert.NotNil(t, l)
	assert.Equal(t, zap.NewAtomicLevelAt(-0).Level(), l.Level())
}
