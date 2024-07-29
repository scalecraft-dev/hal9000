package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"log/slog"

	"github.com/Imagine-Pediatrics/hal/internal/api/config"
	"github.com/Imagine-Pediatrics/hal/internal/api/health"
	"github.com/Imagine-Pediatrics/hal/internal/api/incident"
	"github.com/Imagine-Pediatrics/hal/internal/api/interaction"
	"github.com/joho/godotenv"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

func middleware(slackApi *slack.Client, cfg config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("slackApi", slackApi)

		byteBody, err := ioutil.ReadAll(ctx.Request.Body)
		ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(byteBody))
		if err != nil {
			ctx.JSON(500, gin.H{"error": "Error reading request body"})
		}

		slackTimestamp := ctx.GetHeader("x-slack-request-timestamp")

		hash := hmac.New(sha256.New, []byte(cfg.SlackSigningSecret))
		hash.Write([]byte(fmt.Sprintf("v0:%s:%s", slackTimestamp, string(byteBody))))

		computed := fmt.Sprintf("v0=%s", hex.EncodeToString(hash.Sum(nil)))

		slackSignature := ctx.GetHeader("x-slack-signature")

		if slackSignature == computed {
			ctx.Set("x-valid-slack-request", true)
		} else {
			ctx.Set("x-valid-slack-request", false)
		}
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

	config := config.GetConfig()

	r := gin.Default()
	slackApi := slack.New(config.SlackToken)
	r.Use(middleware(slackApi, *config))

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
