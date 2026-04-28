package postges

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dws-1-2026-green/subscriptions/internal/domain/subscription"
	"github.com/dws-1-2026-green/subscriptions/internal/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PosgresSubscriptionsRepo struct {
	pool *pgxpool.Pool
}

func (r PosgresSubscriptionsRepo) ListBySourceAndType(ctx context.Context, source string, eventType string) ([]subscription.Subscription, error) {
	start := time.Now()
	defer func() {
		metrics.DBQueryDuration.WithLabelValues("postgres").Observe(time.Since(start).Seconds())
	}()

	const q = `
		select
  			id,
  			target_url,
  			http_method,
  			headers
		from
			subscriptions
		where 1=1
			and source = $1
  			and event_type = $2
  			and enabled = true
	`

	rows, err := r.pool.Query(ctx, q, source, eventType)
	if err != nil {
		return nil, fmt.Errorf("query subscriptions: %w", err)
	}
	defer rows.Close()

	out := make([]subscription.Subscription, 0)

	for rows.Next() {
		var s subscription.Subscription
		var headersRaw []byte

		if err := rows.Scan(
			&s.Id,
			&s.DestinnationUrl,
			&s.Method,
			&headersRaw,
		); err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}

		if len(headersRaw) == 0 {
			s.Headers = map[string]string{}
		} else {
			var m map[string]string
			if err := json.Unmarshal(headersRaw, &m); err != nil {
				return nil, fmt.Errorf("unmarshal headers for subscription %s: %w", s.Id, err)
			}

			if m == nil {
				m = map[string]string{}
			}

			s.Headers = m
		}

		out = append(out, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subscriptions: %w", err)
	}

	return out, nil
}

func NewPosgresSubscriptionsRepo(pool *pgxpool.Pool) PosgresSubscriptionsRepo {
	return PosgresSubscriptionsRepo{
		pool: pool,
	}
}
