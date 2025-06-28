package logging

import (
	"github.com/rs/zerolog"
	"os"
	"time"
)

// initLogger initialise loggers configuration
func Init() {
	zerolog.TimeFieldFormat = time.DateTime
	if os.Getenv("APP_ENV") != "production" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel) // Show debug in dev
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel) // Hide debug in prod
	}
}
