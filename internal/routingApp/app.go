package routingApp

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/dws-1-2026-green/subscriptions/internal/adapter/cassandra"
	"github.com/dws-1-2026-green/subscriptions/internal/adapter/kafka"
	"github.com/dws-1-2026-green/subscriptions/internal/adapter/postges"
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
	log.Println("starting routing worker loop...")
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
		log.Printf("connecting to cassandra: hosts=%v keyspace=%s", cfg.CassandraHosts, cfg.CassandraKeyspace)
		cluster := gocql.NewCluster(cfg.CassandraHosts...)
		cluster.Keyspace = cfg.CassandraKeyspace
		cluster.Consistency = gocql.ParseConsistency(cfg.CassandraConsistency)
		session, err := gocqlx.WrapSession(cluster.CreateSession())
		if err != nil {
			return nil, err
		}

		log.Println("successfully connected to cassandra")
		a.closeFuncs = append(a.closeFuncs, closeFunc(session.Close))
		repo = cassandra.NewCassandraSubscriptionsRepo(&session)
	default:
		return nil, errors.New("unknown STORE_BACKEND: " + cfg.StoreBackend)
	}

	handler := routing.NewHandler(repo)

	log.Printf("initializing kafka reader: topic=%s group=%s", cfg.RoutingRequestsTopic, cfg.KafkaGroupID)
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: cfg.KafkaBrokers,
		GroupID: cfg.KafkaGroupID,
		Topic:   cfg.RoutingRequestsTopic,
	})
	a.closeFuncs = append(a.closeFuncs, closeFunc(func() { _ = reader.Close() }))

	log.Printf("initializing kafka writer: topic=%s", cfg.DeliveriesToSendTopic)
	writer := kafkago.NewWriter(kafkago.WriterConfig{
		Brokers:  cfg.KafkaBrokers,
		Topic:    cfg.DeliveriesToSendTopic,
		Balancer: &kafkago.Hash{},
	})
	a.closeFuncs = append(a.closeFuncs, closeFunc(func() { _ = writer.Close() }))

	a.worker = kafka.NewWorker(reader, writer, handler)

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(cfg.MetricsAddr, mux); err != nil {
			log.Printf("metrics server: %v", err)
		}
	}()

	return a, nil
}
