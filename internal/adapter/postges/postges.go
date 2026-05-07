package postges

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dws-1-2026-green/subscriptions/internal/domain/subscription"
	"github.com/dws-1-2026-green/subscriptions/internal/metrics"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PosgresSubscriptionsRepo struct {
	pool *pgxpool.Pool
}

func NewPosgresSubscriptionsRepo(pool *pgxpool.Pool) PosgresSubscriptionsRepo {
	return PosgresSubscriptionsRepo{
		pool: pool,
	}
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
			&s.DestinationUrl,
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

// Create creates a new subscription
func (r PosgresSubscriptionsRepo) Create(ctx context.Context, sub *subscription.Subscription) error {
	start := time.Now()
	defer func() {
		metrics.DBQueryDuration.WithLabelValues("postgres").Observe(time.Since(start).Seconds())
	}()

	const q = `
		INSERT INTO subscriptions (id, source, event_type, target_url, http_method, headers, enabled, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	headersJSON, err := json.Marshal(sub.Headers)
	if err != nil {
		return fmt.Errorf("marshal headers: %w", err)
	}

	_, err = r.pool.Exec(ctx, q,
		sub.Id,
		sub.Source,
		sub.EventType,
		sub.DestinationUrl,
		sub.Method,
		headersJSON,
		sub.Enabled,
		sub.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert subscription: %w", err)
	}

	return nil
}

// GetByID retrieves a subscription by its ID
func (r PosgresSubscriptionsRepo) GetByID(ctx context.Context, id string) (*subscription.Subscription, error) {
	start := time.Now()
	defer func() {
		metrics.DBQueryDuration.WithLabelValues("postgres").Observe(time.Since(start).Seconds())
	}()

	const q = `
		SELECT id, source, event_type, target_url, http_method, headers, enabled, created_at
		FROM subscriptions
		WHERE id = $1
	`

	var sub subscription.Subscription
	var headersRaw []byte

	err := r.pool.QueryRow(ctx, q, id).Scan(
		&sub.Id,
		&sub.Source,
		&sub.EventType,
		&sub.DestinationUrl,
		&sub.Method,
		&headersRaw,
		&sub.Enabled,
		&sub.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, subscription.ErrNotFound
		}
		return nil, fmt.Errorf("query subscription by id: %w", err)
	}

	if len(headersRaw) == 0 {
		sub.Headers = map[string]string{}
	} else {
		if err := json.Unmarshal(headersRaw, &sub.Headers); err != nil {
			return nil, fmt.Errorf("unmarshal headers: %w", err)
		}
	}

	return &sub, nil
}

// List retrieves subscriptions, optionally filtered by source and event type
func (r PosgresSubscriptionsRepo) List(ctx context.Context, source, eventType string) ([]subscription.Subscription, error) {
	start := time.Now()
	defer func() {
		metrics.DBQueryDuration.WithLabelValues("postgres").Observe(time.Since(start).Seconds())
	}()

	var q string
	var args []interface{}

	if source != "" && eventType != "" {
		q = `
			SELECT id, source, event_type, target_url, http_method, headers, enabled, created_at
			FROM subscriptions
			WHERE source = $1 AND event_type = $2
		`
		args = []interface{}{source, eventType}
	} else {
		q = `
			SELECT id, source, event_type, target_url, http_method, headers, enabled, created_at
			FROM subscriptions
		`
		args = nil
	}

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query subscriptions: %w", err)
	}
	defer rows.Close()

	out := make([]subscription.Subscription, 0)

	for rows.Next() {
		var sub subscription.Subscription
		var headersRaw []byte

		if err := rows.Scan(
			&sub.Id,
			&sub.Source,
			&sub.EventType,
			&sub.DestinationUrl,
			&sub.Method,
			&headersRaw,
			&sub.Enabled,
			&sub.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}

		if len(headersRaw) == 0 {
			sub.Headers = map[string]string{}
		} else {
			if err := json.Unmarshal(headersRaw, &sub.Headers); err != nil {
				return nil, fmt.Errorf("unmarshal headers: %w", err)
			}
		}

		out = append(out, sub)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subscriptions: %w", err)
	}

	return out, nil
}

// Update updates an existing subscription
func (r PosgresSubscriptionsRepo) Update(ctx context.Context, sub *subscription.Subscription) error {
	start := time.Now()
	defer func() {
		metrics.DBQueryDuration.WithLabelValues("postgres").Observe(time.Since(start).Seconds())
	}()

	const q = `
		UPDATE subscriptions
		SET target_url = $1, http_method = $2, headers = $3, enabled = $4
		WHERE id = $5
	`

	headersJSON, err := json.Marshal(sub.Headers)
	if err != nil {
		return fmt.Errorf("marshal headers: %w", err)
	}

	result, err := r.pool.Exec(ctx, q,
		sub.DestinationUrl,
		sub.Method,
		headersJSON,
		sub.Enabled,
		sub.Id,
	)
	if err != nil {
		return fmt.Errorf("update subscription: %w", err)
	}

	if result.RowsAffected() == 0 {
		return subscription.ErrNotFound
	}

	return nil
}

// Delete removes a subscription by its ID
func (r PosgresSubscriptionsRepo) Delete(ctx context.Context, id string) error {
	start := time.Now()
	defer func() {
		metrics.DBQueryDuration.WithLabelValues("postgres").Observe(time.Since(start).Seconds())
	}()

	const q = `DELETE FROM subscriptions WHERE id = $1`

	result, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete subscription: %w", err)
	}

	if result.RowsAffected() == 0 {
		return subscription.ErrNotFound
	}

	return nil
}
