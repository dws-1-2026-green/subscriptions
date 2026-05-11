package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/dws-1-2026-green/subscriptions/internal/config"
	"github.com/dws-1-2026-green/subscriptions/internal/logger"
	routingApp "github.com/dws-1-2026-green/subscriptions/internal/routingApp"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("env: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	logger.SetupLogger(&cfg)
	slog.Info("Loaded config", slog.Any("config", cfg))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, err := routingApp.New(ctx, cfg)
	if err != nil {
		log.Fatalf("app: %v", err)
	}
	defer app.Close()

	slog.Info("application initialized, starting worker...")
	if err := app.Run(ctx); err != nil {
		log.Fatalf("worker: %v", err)
	}
}
