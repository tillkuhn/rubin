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
// JSON Schema: https://github.com/cloudevents/spec/blob/main/cloudevents/formats/cloudevents.json
func NewCloudEvent(sourceURI string, eventType string, subject string, data interface{}) (cloudevents.Event, error) {
	event := cloudevents.NewEvent()

	// Set event context, required: ["id", "source", "specversion", "type"]
	// ID  identifies the event.
	event.SetID(uuid.New().String())

	// Identifies the context in which an event happened. URI-reference such as ""//my.apis.com/projects/123/logs
	// Examples "https://github.com/cloudevents", "/sensors/tn-1234567/alerts", "cloudevents/spec/pull/123",
	// "mailto:cncf-wg-serverless@lists.cncf.io", "urn:uuid:6e8bc430-9c3a-11d9-9669-0800200c9a66"
	event.SetSource(sourceURI)

	// Describes the type of event related to the originating occurrence.
	// Examples: "com.github.pull_request.opened" or "com.example.object.deleted.v2"
	// or "google.cloud.pubsub.topic.v1.messagePublished", "google.cloud.storage.object.v1.finalized"
	event.SetType(eventType) // "net.timafe.events.ci.published.v1"

	// Optional "Describes the subject of the event in the context of the event producer (identified by source).",
	// e.g. "newfile.jpg"
	if subject != "" {
		event.SetSubject(subject)
	}

	// "Timestamp of when the occurrence happened. Must adhere to RFC 3339.",
	// make sure we round to .SSS (at least we had an issue when we did not round)
	event.SetTime(time.Now().Round(time.Second))
	_, payload, err := transformPayload(data)

	if err != nil {
		return cloudevents.Event{}, err
	}

	// the actual event payload
	err = event.SetData(cloudevents.ApplicationJSON, payload) // data content type
	return event, err
}
