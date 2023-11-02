package rubin

import (
	"time"

	"go.uber.org/zap"
)

// Client an instance of confluent.Client initialized with the given options
type Client struct {
	options *Options
	logger  zap.SugaredLogger
}

type Event struct {
	Action  string `json:"action,omitempty"`
	Message string `json:"message,omitempty"`
	// Time.MarshalJSON returns
	// "The time is a quoted string in RFC 3339 format, with sub-second precision added if present."
	Time     time.Time `json:"time,omitempty"`
	Source   string    `json:"source,omitempty"`
	EntityID string    `json:"entityId,omitempty"`
}

/*
payload := `{
	 "key": {
	   "type": "BINARY",
	   "data": "Zm9vYmFy"
	 },
	 "value": {
	   "type": "JSON",
	   "data": "Bonjour le monde, de nouveau! Let's go!'"
	 }
	}`
*/
