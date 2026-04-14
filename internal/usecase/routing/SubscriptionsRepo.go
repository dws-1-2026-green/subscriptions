package routing

import (
	"context"

	"github.com/dws-1-2026-green/subscriptions/internal/domain/subscription"
)

type SubscriptionsRepo interface {
	ListBySourceAndType(ctx context.Context, source string, eventType string) ([]subscription.Subscription, error)
}
