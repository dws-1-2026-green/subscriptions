package cassandra

import (
	"context"
	"fmt"
	"time"

	"github.com/dws-1-2026-green/subscriptions/internal/domain/subscription"
	"github.com/dws-1-2026-green/subscriptions/internal/metrics"
	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/qb"
)

type CassandraSubscriptionsRepo struct {
	session *gocqlx.Session
}

func NewCassandraSubscriptionsRepo(session *gocqlx.Session) CassandraSubscriptionsRepo {
	return CassandraSubscriptionsRepo{
		session: session,
	}
}

type subscriptionRow struct {
	Source         string            `db:"source"`
	EventType      string            `db:"event_type"`
	SubscriptionID string            `db:"subscription_id"`
	DestinationURL string            `db:"destination_url"`
	HTTPMethod     string            `db:"http_method"`
	Headers        map[string]string `db:"headers"`
	Enabled        bool              `db:"enabled"`
	CreatedAt      time.Time         `db:"created_at"`
}

func (r CassandraSubscriptionsRepo) ListBySourceAndType(ctx context.Context, source string, eventType string) ([]subscription.Subscription, error) {
	start := time.Now()
	defer func() {
		metrics.DBQueryDuration.WithLabelValues("cassandra").Observe(time.Since(start).Seconds())
	}()

	return r.listBySourceAndType(source, eventType)
}

func (r CassandraSubscriptionsRepo) List(ctx context.Context, source string, eventType string) ([]subscription.Subscription, error) {
	start := time.Now()
	defer func() {
		metrics.DBQueryDuration.WithLabelValues("cassandra").Observe(time.Since(start).Seconds())
	}()

	if source != "" && eventType != "" {
		return r.listBySourceAndType(source, eventType)
	}

	return r.listAll()
}

func (r CassandraSubscriptionsRepo) listBySourceAndType(source, eventType string) ([]subscription.Subscription, error) {
	query := qb.Select("subscriptions").
		Columns("subscription_id", "source", "event_type", "destination_url", "http_method", "headers", "enabled", "created_at").
		Where(qb.Eq("source"), qb.Eq("event_type")).
		Query(*r.session).
		Bind(source, eventType)

	defer query.Release()

	var rows []subscriptionRow
	if err := query.Select(&rows); err != nil {
		return nil, fmt.Errorf("query subscriptions: %w", err)
	}

	return r.rowsToSubscriptions(rows), nil
}

func (r CassandraSubscriptionsRepo) listAll() ([]subscription.Subscription, error) {
	query := qb.Select("subscriptions_by_id").
		Columns("subscription_id", "source", "event_type", "destination_url", "http_method", "headers", "enabled", "created_at").
		Query(*r.session)

	defer query.Release()

	var rows []subscriptionRow
	if err := query.Select(&rows); err != nil {
		return nil, fmt.Errorf("query all subscriptions: %w", err)
	}

	return r.rowsToSubscriptions(rows), nil
}

func (r CassandraSubscriptionsRepo) rowsToSubscriptions(rows []subscriptionRow) []subscription.Subscription {
	out := make([]subscription.Subscription, 0, len(rows))
	for _, row := range rows {
		s := subscription.Subscription{
			Id:             row.SubscriptionID,
			Source:         row.Source,
			EventType:      row.EventType,
			DestinationUrl: row.DestinationURL,
			Method:         row.HTTPMethod,
			Headers:        row.Headers,
			Enabled:        row.Enabled,
			CreatedAt:      row.CreatedAt,
		}

		if s.Headers == nil {
			s.Headers = map[string]string{}
		}

		out = append(out, s)
	}
	return out
}

func (r CassandraSubscriptionsRepo) Create(ctx context.Context, sub *subscription.Subscription) error {
	start := time.Now()
	defer func() {
		metrics.DBQueryDuration.WithLabelValues("cassandra").Observe(time.Since(start).Seconds())
	}()

	batch := r.session.NewBatch(gocql.LoggedBatch)

	batch.Query(
		`INSERT INTO subscriptions (source, event_type, subscription_id, destination_url, http_method, headers, enabled, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		sub.Source, sub.EventType, sub.Id, sub.DestinationUrl, sub.Method, sub.Headers, sub.Enabled, sub.CreatedAt,
	)

	batch.Query(
		`INSERT INTO subscriptions_by_id (subscription_id, source, event_type, destination_url, http_method, headers, enabled, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		sub.Id, sub.Source, sub.EventType, sub.DestinationUrl, sub.Method, sub.Headers, sub.Enabled, sub.CreatedAt,
	)

	if err := r.session.Session.ExecuteBatch(batch); err != nil {
		return fmt.Errorf("batch insert subscription: %w", err)
	}

	return nil
}

func (r CassandraSubscriptionsRepo) GetByID(ctx context.Context, id string) (*subscription.Subscription, error) {
	start := time.Now()
	defer func() {
		metrics.DBQueryDuration.WithLabelValues("cassandra").Observe(time.Since(start).Seconds())
	}()

	query := qb.Select("subscriptions_by_id").
		Columns("subscription_id", "source", "event_type", "destination_url", "http_method", "headers", "enabled", "created_at").
		Where(qb.Eq("subscription_id")).
		Query(*r.session).
		Bind(id)

	defer query.Release()

	var row subscriptionRow
	if err := query.Get(&row); err != nil {
		if err == gocql.ErrNotFound {
			return nil, subscription.ErrNotFound
		}
		return nil, fmt.Errorf("query subscription by id: %w", err)
	}

	return &subscription.Subscription{
		Id:             row.SubscriptionID,
		Source:         row.Source,
		EventType:      row.EventType,
		DestinationUrl: row.DestinationURL,
		Method:         row.HTTPMethod,
		Headers:        row.Headers,
		Enabled:        row.Enabled,
		CreatedAt:      row.CreatedAt,
	}, nil
}

func (r CassandraSubscriptionsRepo) Update(ctx context.Context, sub *subscription.Subscription) error {
	start := time.Now()
	defer func() {
		metrics.DBQueryDuration.WithLabelValues("cassandra").Observe(time.Since(start).Seconds())
	}()

	batch := r.session.NewBatch(gocql.LoggedBatch)

	batch.Query(
		`UPDATE subscriptions SET destination_url = ?, http_method = ?, headers = ?, enabled = ? WHERE source = ? AND event_type = ? AND subscription_id = ?`,
		sub.DestinationUrl, sub.Method, sub.Headers, sub.Enabled, sub.Source, sub.EventType, sub.Id,
	)

	batch.Query(
		`UPDATE subscriptions_by_id SET destination_url = ?, http_method = ?, headers = ?, enabled = ? WHERE subscription_id = ?`,
		sub.DestinationUrl, sub.Method, sub.Headers, sub.Enabled, sub.Id,
	)

	if err := r.session.Session.ExecuteBatch(batch); err != nil {
		return fmt.Errorf("batch update subscription: %w", err)
	}

	return nil
}

func (r CassandraSubscriptionsRepo) Delete(ctx context.Context, id string) error {
	start := time.Now()
	defer func() {
		metrics.DBQueryDuration.WithLabelValues("cassandra").Observe(time.Since(start).Seconds())
	}()

	sub, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	batch := r.session.NewBatch(gocql.LoggedBatch)

	batch.Query(
		`DELETE FROM subscriptions WHERE source = ? AND event_type = ? AND subscription_id = ?`,
		sub.Source, sub.EventType, id,
	)

	batch.Query(
		`DELETE FROM subscriptions_by_id WHERE subscription_id = ?`,
		id,
	)

	if err := r.session.Session.ExecuteBatch(batch); err != nil {
		return fmt.Errorf("batch delete subscription: %w", err)
	}

	return nil
}
