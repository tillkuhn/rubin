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
		// ProducerClientID:   "",
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
	_ = os.Setenv(prefix+"_BOOTSTRAP_SERVERS", "boot.strap.io")
	o, err = NewOptionsFromEnv()
	assert.NoError(t, err)
	assert.Equal(t, "key-west", o.ConsumerAPIKey)
	assert.Equal(t, "boot.strap.io", o.BootstrapServers)
	assert.Contains(t, o.String(), "boot.strap.io")
}

func TestOptionsError(t *testing.T) {
	defer os.Clearenv()
	_ = os.Setenv(prefix+"_CONSUMER_MAX_RECEIVE", "this is not a number")
	_, err := NewOptionsFromEnv()
	assert.ErrorContains(t, err, "invalid syntax") // envconfig.Process: strconv.ParseInt: ...
}
