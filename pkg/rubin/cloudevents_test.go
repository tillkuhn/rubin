package rubin

import (
	"encoding/json"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/stretchr/testify/assert"
)

type testData struct {
	Action  string `json:"action"`
	Message string `json:"message"`
}

func TestNewCloudEvent(t *testing.T) {
	m := map[string]string{"action": "test/me", "message": "test output"}
	event, err := NewCloudEvent("//testing/event", "", m)
	subject := "my.subject"
	event.SetSubject(subject)
	assert.NoError(t, err)
	assert.NotEmpty(t, event.ID())
	bytes, err := json.MarshalIndent(event, "", "  ")
	assert.NoError(t, err)
	assert.True(t, len(bytes) > 0)
	assert.Contains(t, string(bytes), subject)
	// t.Log("\n" + string(bytes)) // uncomment for debug

	// try to unmarshal the previous string map into a struct with the same field
	var m2 testData
	err = event.DataAs(&m2)
	assert.NoError(t, err)
	assert.Equal(t, "test/me", m2.Action) // should work :-)
	assert.NoError(t, err)
	assert.True(t, len(bytes) > 0)
	// t.Log("\nMAP EVENT:" + string(bytes))

	jsonString := `{"message":"Hello Franz!"}`
	event, err = NewCloudEvent("//testing/event", "your.subject", jsonString)
	assert.NoError(t, err)
	bytes, err = json.MarshalIndent(event, "", "  ")
	assert.NoError(t, err)
	assert.True(t, len(bytes) > 0)
	assert.Equal(t, event.DataContentType(), cloudevents.ApplicationJSON)
	// t.Log("\nJSON STRING:" + string(bytes))

	// test with no data (which is optional)
	event, err = NewCloudEvent("//testing/event", "your.subject", nil)
	assert.NoError(t, err)

	// test with simple string
	event, err = NewCloudEvent("//testing/event", "your.subject", "this is a string")
	assert.NoError(t, err)
	assert.Equal(t, event.DataContentType(), cloudevents.TextPlain)
}
