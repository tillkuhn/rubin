package rubin

import (
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

// NewCloudEvent returns a cloud event initialized with default data and payload based on data map,
// source and type are required arguments
// data could be data map[string]string or some struct, as long as it can be JSON-serialized
//
//	event, err := NewCloudEvent("//testing/event", "my.event.type", data)
//	subject := "my.subject"
//
// Inspired by official Go SDK for CloudEvents https://github.com/cloudevents/sdk-go
// Specs:
// - Detailed attribute descriptions: https://github.com/cloudevents/spec/blob/v1.0/spec.md#event-format
// - JSON specific https://github.com/cloudevents/spec/blob/main/cloudevents/formats/json-format.md
// - Actual JSON Schema: https://github.com/cloudevents/spec/blob/main/cloudevents/formats/cloudevents.json
//
// Real World Examples:
// - https://cloud.google.com/eventarc/docs/workflows/cloudevents (Google APIs)
// - https://github.com/googleapis/google-cloudevents/tree/main/examples/structured (Google APIs)
// - https://docs.dapr.io/developing-applications/building-blocks/pubsub/pubsub-cloudevents/ (Dapr)

func NewCloudEvent(sourceURI string, eventType string, data interface{}) (cloudevents.Event, error) {
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
	// "com.dapr.event.sent", "net.timafe.event.app.started", "net.timafe.event.entity.created"
	event.SetType(eventType) //
	// "Timestamp of when the occurrence happened. Must adhere to RFC 3339.",
	// make sure we round to .SSS (at least we had an issue when we did not round)
	event.SetTime(time.Now().Round(time.Second))

	// Optional Subjects "Describes the subject of the event in the context of the event producer (identified by source)."
	// the receiver is expected to set Subject and other optional attributes on the returned Event struct

	// Optional data with some JSON String serialization magic
	if data != nil {
		vType, payload, err := transformPayload(data)

		if err != nil {
			return event, err
		}

		cType := cloudevents.ApplicationJSON
		if vType == "STRING" {
			cType = cloudevents.TextPlain
		}
		err = event.SetData(cType, payload) // data content type
		if err != nil {
			return event, err
		}
	}
	return event, nil
}

// Old Event Format
// type Event struct {
//	Action  string `json:"action,omitempty"`
//	Message string `json:"message,omitempty"`
//	// the time is  a quoted string in RFC 3339 format, with sub-second precision added if present."
//	Time     time.Time `json:"time,omitempty"`
//	Source   string    `json:"source,omitempty"`
//	EntityID string    `json:"entityId,omitempty"`
//}
