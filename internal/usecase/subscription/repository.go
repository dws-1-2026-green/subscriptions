package subscription

import (
	"context"

	"github.com/dws-1-2026-green/subscriptions/internal/domain/subscription"
)

// Repository defines the interface for subscription storage operations
type Repository interface {
	// Create creates a new subscription
	Create(ctx context.Context, sub *subscription.Subscription) error

	// GetByID retrieves a subscription by its ID
	GetByID(ctx context.Context, id string) (*subscription.Subscription, error)

	// List retrieves subscriptions, optionally filtered by source and event type
	List(ctx context.Context, source, eventType string) ([]subscription.Subscription, error)

	// Update updates an existing subscription
	Update(ctx context.Context, sub *subscription.Subscription) error

	// Delete removes a subscription by its ID
	Delete(ctx context.Context, id string) error
}
