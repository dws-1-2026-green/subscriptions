package routingApp

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/dws-1-2026-green/subscriptions/internal/adapter/cassandra"
	"github.com/dws-1-2026-green/subscriptions/internal/adapter/kafka"
	"github.com/dws-1-2026-green/subscriptions/internal/adapter/postges"
	"github.com/dws-1-2026-green/subscriptions/internal/cache"
	"github.com/dws-1-2026-green/subscriptions/internal/config"
	"github.com/dws-1-2026-green/subscriptions/internal/usecase/routing"
	"github.com/dws-1-2026-green/subscriptions/internal/worker"
	"github.com/gocql/gocql"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/scylladb/gocqlx/v2"

	kafkago "github.com/segmentio/kafka-go"
)

type closeFunc func()

type RoutingApp struct {
	cfg config.Config

	closeFuncs []closeFunc

	worker worker.Worker
}

func (a *RoutingApp) Run(ctx context.Context) error {
	slog.Info("Starting routing worker loop")
	return a.worker.Run(ctx)
}

func (a *RoutingApp) Close() {
	for _, close := range a.closeFuncs {
		close()
	}
}

func New(ctx context.Context, cfg config.Config) (*RoutingApp, error) {
	a := &RoutingApp{
		cfg:        cfg,
		closeFuncs: make([]closeFunc, 0, 4),
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
	case "cassandra":
		cluster := gocql.NewCluster(cfg.CassandraHosts...)
		cluster.Keyspace = cfg.CassandraKeyspace
		cluster.Consistency = gocql.ParseConsistency(cfg.CassandraConsistency)
		if cfg.CassandraUser != "" {
			cluster.Authenticator = gocql.PasswordAuthenticator{
				Username: cfg.CassandraUser,
				Password: cfg.CassandraPassword,
			}
		}
		session, err := gocqlx.WrapSession(cluster.CreateSession())
		if err != nil {
			return nil, err
		}
		a.closeFuncs = append(a.closeFuncs, closeFunc(session.Close))
		repo = cassandra.NewCassandraSubscriptionsRepo(&session)
	default:
		return nil, errors.New("unknown STORE_BACKEND: " + cfg.StoreBackend)
	}

	slog.Info("Wrapping repo with TTL cache", slog.Duration("ttl", cfg.CacheTTL))
	cachedRepo := cache.NewCachedRepo(repo, cfg.CacheTTL)

	handler := routing.NewHandler(cachedRepo)

	slog.Info("Initializing kafka reader", slog.String("topic", cfg.RoutingRequestsTopic), slog.String("group", cfg.KafkaGroupID))
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: cfg.KafkaBrokers,
		GroupID: cfg.KafkaGroupID,
		Topic:   cfg.RoutingRequestsTopic,
	})
	a.closeFuncs = append(a.closeFuncs, closeFunc(func() { _ = reader.Close() }))

	slog.Info("Initializing kafka writer", slog.String("topic", cfg.DeliveriesToSendTopic))
	writer := kafkago.NewWriter(kafkago.WriterConfig{
		Brokers:      cfg.KafkaBrokers,
		Topic:        cfg.DeliveriesToSendTopic,
		Balancer:     &kafkago.Hash{},
		BatchTimeout: time.Millisecond,
	})
	a.closeFuncs = append(a.closeFuncs, closeFunc(func() { _ = writer.Close() }))

	slog.Info("Initializing kafka worker", slog.Int("concurrency", cfg.WorkerConcurrency))
	a.worker = kafka.NewWorker(reader, writer, handler, cfg.WorkerConcurrency)

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(cfg.MetricsAddr, mux); err != nil {
			slog.Error("Metrics server error", slog.Any("err", err))
		}
	}()

	return a, nil
}
