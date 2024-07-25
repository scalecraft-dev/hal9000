package main

import (
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/Imagine-Pediatrics/hal/internal/api/health"
	"github.com/Imagine-Pediatrics/hal/internal/api/incident"
	"github.com/Imagine-Pediatrics/hal/internal/api/interaction"
	"github.com/joho/godotenv"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

func middleware(slackApi *slack.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("slackApi", slackApi)
		ctx.Next()
	}
}

func init() {
	err := godotenv.Load()

	if err != nil {
		slog.Warn("Error loading .env file")
	}
}

func main() {
	flag.Parse()

	slackToken, ok := os.LookupEnv("SLACK_TOKEN")

	if !ok {
		log.Fatal("SLACK_TOKEN environment variable not set")
	}

	r := gin.Default()
	slackApi := slack.New(slackToken)
	r.Use(middleware(slackApi))

	err := r.SetTrustedProxies(nil)

	if err != nil {
		log.Fatalf("error setting trusted proxies: %v", err)
	}

	r.GET("/health", health.GetHandler)
	r.POST("/incident", incident.PostHandler(slackApi))
	r.POST("/interaction", interaction.PostHandler(slackApi))

	err = r.Run(":50051")

	if err != nil {
		log.Fatalf("could not start server %v", err)
	}
}
