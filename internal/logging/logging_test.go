package logging

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestApplyLogLevel(t *testing.T) {
	tests := []struct {
		name      string
		verbosity string
		want      string
	}{
		{"valid_debug", "debug", zerolog.DebugLevel.String()},
		{"valid_info", "info", zerolog.InfoLevel.String()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyLogLevel(tt.verbosity)
			if got != tt.want {
				t.Errorf("ApplyLogLevel(%q) = %q, want %q", tt.verbosity, got, tt.want)
			}
		})
	}
}
