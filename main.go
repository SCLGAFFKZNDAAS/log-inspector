package main

import (
	"log-inspector/llm"
	"log-inspector/logger"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Load envfile
	if os.Getenv("GO_ENV") != "production" {
		godotenv.Load(".env")
	}

	if os.Getenv("LOKI_URL") == "" {
		panic("LOKI_URL is not set")
	}
	if os.Getenv("LLM_URL") == "" {
		panic("LLM_URL is not set")
	}
	if os.Getenv("MODEL_NAME") == "" {
		panic("MODEL_NAME is not set")
	}
	if os.Getenv("SERVICES_TO_SEARCH") == "" {
		panic("SERVICES_TO_SEARCH is not set")
	}
	TIME_BETWEEN_CHECKS := os.Getenv("TIME_BETWEEN_CHECKS")
	if TIME_BETWEEN_CHECKS == "" {
		panic("TIME_BETWEEN_CHECKS is not set")
	}
	parsedTime, timeErr := time.ParseDuration(TIME_BETWEEN_CHECKS)
	if timeErr != nil {
		panic("Invalid TIME_BETWEEN_CHECKS format, should be a duration like '5m' or '1h'")
	}

	logger.InitLogger()

	var lastPeriodicCheckResult llm.FinalJSONResponse
	var err error
	lastPeriodicCheckTime := time.Now()

	go func() {
		for {
			timeSinceLastCheck := time.Since(lastPeriodicCheckTime)
			if timeSinceLastCheck > parsedTime {
				log.Error().Str("src", "program").Dur("time_since_last_check", timeSinceLastCheck).Msg("Periodic check is taking longer than the specified interval")
			}
			time.Sleep(1 * time.Minute)
		}
	}()

	for {
		t0 := time.Now()
		log.Info().Str("src", "program").Msg("Starting periodic check")
		lastPeriodicCheckResult, err = PerformPeriodicCheck(lastPeriodicCheckResult)
		if err != nil {
			log.Error().Str("src", "program").Err(err).Msg("Error performing periodic check")
		}
		lastPeriodicCheckTime = time.Now()
		timeToWait := parsedTime - time.Since(t0)
		if timeToWait < 0 {
			log.Error().Str("src", "program").Msg("Periodic check took longer than the specified interval, starting next check immediately")
			continue
		}
		log.Info().Str("src", "program").Dur("time_to_wait", timeToWait).Msg("Sleeping until next periodic check")
		time.Sleep(timeToWait)
	}
}

func PerformPeriodicCheck(lastPeriodicCheckResult llm.FinalJSONResponse) (llm.FinalJSONResponse, error) {
	lastPeriodicCheckResult, err := llm.PeriodicCheck(lastPeriodicCheckResult)
	if err != nil {
		log.Error().Str("src", "program").Err(err).Msg("Error in periodic check")
	}

	for _, logEntry := range lastPeriodicCheckResult.Logs {
		var logEvent *zerolog.Event
		switch logEntry.Level {
		case "critical":
			logEvent = log.Log().Str("level", "critical")
		case "error":
			logEvent = log.Error()
		case "warning":
			logEvent = log.Warn()
		default:
			logEvent = log.Info()
		}

		logEvent.Str("src", "report").
			Str("type", logEntry.Type).
			Str("time", logEntry.Time).
			Str("original_log", logEntry.OriginalLog).
			Str("llm_investigation", logEntry.LLMInvestigation).
			Str("llm_suggested_action", logEntry.LLMSuggestedAction).
			Msg(logEntry.Message)
	}

	return lastPeriodicCheckResult, err
}
