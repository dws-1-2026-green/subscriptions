package cache

import (
	"context"
	"sync"
	"time"

	"github.com/dws-1-2026-green/subscriptions/internal/domain/subscription"
	"github.com/dws-1-2026-green/subscriptions/internal/metrics"
	"github.com/dws-1-2026-green/subscriptions/internal/usecase/routing"
)

type entry struct {
	subs      []subscription.Subscription
	expiresAt time.Time
}

// CachedRepo is a TTL cache decorator over SubscriptionsRepo.
// Safe for concurrent use.
type CachedRepo struct {
	inner routing.SubscriptionsRepo
	ttl   time.Duration
	mu    sync.RWMutex
	data  map[string]entry
}

func NewCachedRepo(inner routing.SubscriptionsRepo, ttl time.Duration) *CachedRepo {
	return &CachedRepo{
		inner: inner,
		ttl:   ttl,
		data:  make(map[string]entry),
	}
}

func (c *CachedRepo) ListBySourceAndType(ctx context.Context, source, eventType string) ([]subscription.Subscription, error) {
	// \x00 is safe as a separator — source/eventType are user-defined text that cannot contain null bytes
	key := source + "\x00" + eventType
	now := time.Now()

	c.mu.RLock()
	e, ok := c.data[key]
	c.mu.RUnlock()

	if ok && now.Before(e.expiresAt) {
		metrics.CacheHitsTotal.Inc()
		return e.subs, nil
	}

	metrics.CacheMissesTotal.Inc()

	subs, err := c.inner.ListBySourceAndType(ctx, source, eventType)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.data[key] = entry{subs: subs, expiresAt: now.Add(c.ttl)}
	c.mu.Unlock()

	return subs, nil
}
