package subscription

import (
	"context"
	"time"

	"github.com/dws-1-2026-green/subscriptions/internal/domain/subscription"
	"github.com/google/uuid"
)

// Service provides business logic for subscription management
type Service struct {
	repo Repository
}

// NewService creates a new subscription service
func NewService(repo Repository) Service {
	return Service{repo: repo}
}

// Create creates a new subscription
func (s Service) Create(ctx context.Context, req *subscription.CreateRequest) (*subscription.Subscription, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	sub := &subscription.Subscription{
		Id:             uuid.New().String(),
		Source:         req.Source,
		EventType:      req.EventType,
		DestinationUrl: req.DestinationUrl,
		Method:         req.Method,
		Headers:        req.Headers,
		Enabled:        true,
		CreatedAt:      time.Now(),
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

// GetByID retrieves a subscription by its ID
func (s Service) GetByID(ctx context.Context, id string) (*subscription.Subscription, error) {
	return s.repo.GetByID(ctx, id)
}

// List retrieves subscriptions, optionally filtered by source and event type
func (s Service) List(ctx context.Context, source, eventType string) ([]subscription.Subscription, error) {
	return s.repo.List(ctx, source, eventType)
}

// Update updates an existing subscription
func (s Service) Update(ctx context.Context, id string, req *subscription.UpdateRequest) (*subscription.Subscription, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.DestinationUrl != "" {
		existing.DestinationUrl = req.DestinationUrl
	}
	if req.Method != "" {
		existing.Method = req.Method
	}
	if req.Headers != nil {
		existing.Headers = req.Headers
	}
	if req.Enabled != nil {
		existing.Enabled = *req.Enabled
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// Delete removes a subscription by its ID
func (s Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
