package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dws-1-2026-green/subscriptions/internal/config"
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

	log.Printf("configuration loaded: store=%s", cfg.StoreBackend)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, err := routingApp.New(ctx, cfg)
	if err != nil {
		log.Fatalf("app: %v", err)
	}
	defer app.Close()

	log.Println("application initialized, starting worker...")
	if err := app.Run(ctx); err != nil {
		log.Fatalf("worker: %v", err)
	}
}
