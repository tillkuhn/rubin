package polly

import (
	"os"
	"strings"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	var err error
	o := &Options{
		ProducerClientID:   "",
		ConsumerAPIKey:     "",
		ConsumerAPISecret:  "",
		ConsumerGroupID:    "",
		ConsumerMaxReceive: 0,
		ConsumerStartLast:  false,
		Debug:              false,
		BootstrapEndpoint:  "",
	}
	assert.Equal(t, kafka.FirstOffset, o.StartOffset())
	o.ConsumerStartLast = true
	assert.Equal(t, kafka.LastOffset, o.StartOffset())
	prefix := strings.ToUpper(envconfigDefaultPrefix)
	_ = os.Setenv(prefix+"_CONSUMER_API_KEY", "key-west")
	o, err = NewOptionsFromEnv()
	assert.NoError(t, err)
	assert.Equal(t, "key-west", o.ConsumerAPIKey)
}
