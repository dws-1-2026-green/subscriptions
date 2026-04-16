package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dws-1-2026-green/subscriptions/internal/app"
	"github.com/dws-1-2026-green/subscriptions/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("env: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatalf("app: %v", err)
	}
	defer app.Close()

	if err := app.Run(ctx); err != nil {
		log.Fatalf("worker: %v", err)
	}
}
