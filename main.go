package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Imagine-Pediatrics/hal/internal"
	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

func main() {
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := internal.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	slackClient := slack.New(cfg.SlackToken)
	slackService := internal.NewSlackService(slackClient, cfg)
	incidentService := internal.NewIncidentService(slackService, cfg)

	router := gin.Default()
	router.Use(internal.SlackAuthMiddleware(cfg.SlackSigningSecret))
	internal.RegisterRoutes(router, incidentService, slackService)

	srv := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: router,
	}

	if srv.Addr == ":" {
		srv.Addr = ":50051"
	}

	go func() {
		slog.Info("Starting server", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		log.Fatal("Server forced to shutdown:", err)
	}

	slog.Info("Server exited")
}
