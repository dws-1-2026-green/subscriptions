package subscription

import (
	"context"
	"time"

	"github.com/dws-1-2026-green/subscriptions/internal/domain/subscription"
	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return Service{repo: repo}
}

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

func (s Service) GetByID(ctx context.Context, id string) (*subscription.Subscription, error) {
	return s.repo.GetByID(ctx, id)
}

func (s Service) List(ctx context.Context, source, eventType string) ([]subscription.Subscription, error) {
	return s.repo.List(ctx, source, eventType)
}

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

func (s Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
