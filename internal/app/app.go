package app

import (
	"context"
	"errors"

	"github.com/dws-1-2026-green/subscriptions/internal/adapter/kafka"
	"github.com/dws-1-2026-green/subscriptions/internal/adapter/postges"
	"github.com/dws-1-2026-green/subscriptions/internal/config"
	"github.com/dws-1-2026-green/subscriptions/internal/usecase/routing"
	"github.com/dws-1-2026-green/subscriptions/internal/worker"
	"github.com/jackc/pgx/v5/pgxpool"
)

type closeFunc func()

type App struct {
	cfg config.Config

	closeFuncs []closeFunc

	worker worker.Worker
}

func (a *App) Run(ctx context.Context) error {
	return a.worker.Run(ctx)
}

func (a *App) Close() {
	for _, close := range a.closeFuncs {
		close()
	}
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	a := &App{
		cfg:        cfg,
		closeFuncs: make([]closeFunc, 0),
	}

	var repo routing.SubscriptionsRepo
	switch cfg.StoreBackend {
	case "postgres":
		pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, err
		}
		a.closeFuncs = append(a.closeFuncs, closeFunc(pool.Close))
		repo = postges.NewPosgresSubscriptionsRepo(pool)
	default:
		return nil, errors.New("unknown STORE_BACKEND: " + cfg.StoreBackend)
	}

	handler := routing.NewHandler(repo)
	a.worker = kafka.NewWorker(nil, nil, handler)

	return a, nil
}
