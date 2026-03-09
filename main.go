package main

import (
	"fmt"
	"log-inspector/loki"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load envfile
	GO_ENV := os.Getenv("GO_ENV")
	if GO_ENV != "production" {
		godotenv.Load(".env")
	}

	LOKI_URL := os.Getenv("LOKI_URL")
	if LOKI_URL == "" {
		panic("LOKI_URL is not set")
	}

	lokiResp, err := loki.QueryLoki(loki.LokiQuery{
		Query: `{app="nextcloud"} |= ""`,
		Start: time.Now().Add(-5 * time.Minute),
		End:   time.Now(),
		Limit: 100,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Loki response: %+v\n", lokiResp)
}
