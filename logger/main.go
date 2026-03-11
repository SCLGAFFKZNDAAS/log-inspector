package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func InitLogger() {
	zerolog.TimeFieldFormat = time.RFC3339

	if os.Getenv("GO_ENV") == "prod" {
		// Log to docker log collector (stdout)
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	} else {
		// Local/dev → log to console with colors
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		log.Logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
		log.Info().Msg("Logging initialized in console mode")
	}
}
