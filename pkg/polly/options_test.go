package polly

import (
	"os"
	"strings"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

var prefix = strings.ToUpper(envconfigDefaultPrefix)

func TestOptions(t *testing.T) {
	defer os.Clearenv()
	var err error
	o := &Options{
		ProducerClientID:   "",
		ConsumerAPIKey:     "",
		ConsumerAPISecret:  "",
		ConsumerGroupID:    "",
		ConsumerMaxReceive: 0,
		ConsumerStartLast:  false,
		Debug:              false,
		BootstrapServers:   "",
	}
	assert.Equal(t, kafka.FirstOffset, o.StartOffset())
	o.ConsumerStartLast = true
	assert.Equal(t, kafka.LastOffset, o.StartOffset())
	_ = os.Setenv(prefix+"_CONSUMER_API_KEY", "key-west")
	o, err = NewOptionsFromEnv()
	assert.NoError(t, err)
	assert.Equal(t, "key-west", o.ConsumerAPIKey)
}

func TestOptionsError(t *testing.T) {
	defer os.Clearenv()
	_ = os.Setenv(prefix+"_CONSUMER_MAX_RECEIVE", "this is not a number")
	_, err := NewOptionsFromEnv()
	assert.ErrorContains(t, err, "invalid syntax") // envconfig.Process: strconv.ParseInt: ...
}
