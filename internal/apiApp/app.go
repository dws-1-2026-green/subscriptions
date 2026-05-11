package apiApp

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/dws-1-2026-green/subscriptions/internal/adapter/cassandra"
	"github.com/dws-1-2026-green/subscriptions/internal/adapter/postges"
	"github.com/dws-1-2026-green/subscriptions/internal/config"
	"github.com/dws-1-2026-green/subscriptions/internal/httpapi"
	subscriptionusecase "github.com/dws-1-2026-green/subscriptions/internal/usecase/subscription"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gocql/gocql"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/scylladb/gocqlx/v2"
)

type closeFunc func()

type ApiApp struct {
	cfg        config.Config
	server     *http.Server
	closeFuncs []closeFunc
}

func (a *ApiApp) Run(ctx context.Context) error {
	slog.Info("Starting HTTP API server", slog.String("addr", a.cfg.HTTPAddr))
	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (a *ApiApp) Close() {
	slog.Info("Shutting down HTTP API server")
	if err := a.server.Shutdown(context.Background()); err != nil {
		slog.Error("Server shutdown error", slog.Any("err", err))
	}
	for _, close := range a.closeFuncs {
		close()
	}
	slog.Info("HTTP API server stopped")
}

func New(ctx context.Context, cfg config.Config) (*ApiApp, error) {
	a := &ApiApp{
		cfg:        cfg,
		closeFuncs: make([]closeFunc, 0, 2),
	}

	// Initialize repository based on store backend
	var repo subscriptionusecase.Repository
	switch cfg.StoreBackend {
	case "postgres":
		pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, err
		}
		a.closeFuncs = append(a.closeFuncs, closeFunc(pool.Close))
		repo = postges.NewPosgresSubscriptionsRepo(pool)
	case "cassandra":
		cluster := gocql.NewCluster(cfg.CassandraHosts...)
		cluster.Keyspace = cfg.CassandraKeyspace
		cluster.Consistency = gocql.ParseConsistency(cfg.CassandraConsistency)
		session, err := gocqlx.WrapSession(cluster.CreateSession())
		if err != nil {
			return nil, err
		}
		a.closeFuncs = append(a.closeFuncs, closeFunc(session.Close))
		repo = cassandra.NewCassandraSubscriptionsRepo(&session)
	default:
		return nil, errors.New("unknown STORE_BACKEND: " + cfg.StoreBackend)
	}

	// Initialize service and handler
	service := subscriptionusecase.NewService(repo)
	handler := httpapi.NewHandler(service)

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Register routes
	handler.RegisterRoutes(r)

	// Metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Create server
	a.server = &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: r,
	}

	return a, nil
}
