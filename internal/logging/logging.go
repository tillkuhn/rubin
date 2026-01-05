package logging

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ApplyLogLevel adapt global log level depending on verbosity argument
// returns current level
func ApplyLogLevel(verbosity string) string {
	if l, err := zerolog.ParseLevel(verbosity); err == nil {
		zerolog.SetGlobalLevel(l)
	} else {
		log.Error().Msgf("Invalid level %s, skip apply: %v", verbosity, err)
	}
	return zerolog.GlobalLevel().String()
}
