package cassandra

import (
	"context"
	"fmt"
	"time"

	"github.com/dws-1-2026-green/subscriptions/internal/domain/subscription"
	"github.com/dws-1-2026-green/subscriptions/internal/metrics"
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
	CreatedAt      string            `db:"created_at"`
}

func (r CassandraSubscriptionsRepo) ListBySourceAndType(ctx context.Context, source string, eventType string) ([]subscription.Subscription, error) {
	start := time.Now()
	defer func() {
		metrics.DBQueryDuration.WithLabelValues("cassandra").Observe(time.Since(start).Seconds())
	}()

	query := qb.Select("subscriptions").
		Columns("subscription_id", "destination_url", "http_method", "headers", "enabled").
		Where(qb.Eq("source"), qb.Eq("event_type")).
		Query(*r.session).
		Bind(source, eventType)

	defer query.Release()

	var rows []subscriptionRow
	if err := query.Select(&rows); err != nil {
		return nil, fmt.Errorf("query subscriptions: %w", err)
	}

	out := make([]subscription.Subscription, 0, len(rows))
	for _, row := range rows {
		if !row.Enabled {
			continue
		}

		s := subscription.Subscription{
			Id:             row.SubscriptionID,
			DestinationUrl: row.DestinationURL,
			Method:         row.HTTPMethod,
			Headers:        row.Headers,
		}

		if s.Headers == nil {
			s.Headers = map[string]string{}
		}

		out = append(out, s)
	}

	return out, nil
}
