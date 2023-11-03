package rubin

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testData struct {
	Action  string `json:"action"`
	Message string `json:"message"`
}

func TestNewCloudEvent(t *testing.T) {
	m := map[string]string{"action": "test/me", "message": "test output"}
	event, err := NewCloudEvent("//testing/event", m)
	assert.NoError(t, err)
	assert.NotEmpty(t, event.ID())
	bytes, err := json.MarshalIndent(event, "", "  ")
	assert.NoError(t, err)
	assert.True(t, len(bytes) > 0)
	t.Log("\n" + string(bytes))

	// try to unmarshal the previous string map into a struct with the same field
	var m2 testData
	err = event.DataAs(&m2)
	assert.NoError(t, err)
	assert.Equal(t, "test/me", m2.Action) // should work :-)
	assert.NoError(t, err)
	assert.True(t, len(bytes) > 0)
	// t.Log("\nMAP EVENT:" + string(bytes))

	jsonString := `{"message":"Hello Franz!"}`
	event, err = NewCloudEvent("//testing/event", jsonString)
	assert.NoError(t, err)
	bytes, err = json.MarshalIndent(event, "", "  ")
	assert.NoError(t, err)
	assert.True(t, len(bytes) > 0)
	// t.Log("\nJSON STRING:" + string(bytes))
}
