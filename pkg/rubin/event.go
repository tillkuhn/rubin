package rubin

import (
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

// Old Event Format
// type Event struct {
//	Action  string `json:"action,omitempty"`
//	Message string `json:"message,omitempty"`
//	// Time.MarshalJSON returns
//	// Time is  a quoted string in RFC 3339 format, with sub-second precision added if present."
//	Time     time.Time `json:"time,omitempty"`
//	Source   string    `json:"source,omitempty"`
//	EntityID string    `json:"entityId,omitempty"`
//}

// NewCloudEvent returns a cloud event initialized with default data and payload based on data map,
// inspired by official Go SDK for CloudEvents https://github.com/cloudevents/sdk-go
//
// data could be data map[string]string or some struct, as long as it can be JSON-serialized
//
// Spec: https://github.com/cloudevents/spec/blob/main/cloudevents/formats/json-format.md
// Real World Examples: https://cloud.google.com/eventarc/docs/workflows/cloudevents
func NewCloudEvent(sourceURI string, data interface{}) (cloudevents.Event, error) {
	event := cloudevents.NewEvent()
	event.SetID(uuid.New().String())

	// The Source of the event,  URI-reference such as ""//my.apis.com/projects/123/logs
	event.SetSource(sourceURI)

	// The type of event data "ny.domain.v1.dataWritten",
	event.SetType("example.type")

	// subject: attribute specific to the event type.
	// event.SetSubject()

	// event time = current time, make sure we round to .SSS
	event.SetTime(time.Now().Round(time.Second))
	_, payload, err := transformPayload(data)

	if err != nil {
		return cloudevents.Event{}, err
	}

	err = event.SetData(cloudevents.ApplicationJSON, payload) // data content type
	return event, err
}
