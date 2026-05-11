package subscription

import (
	"context"
	"log/slog"
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
		slog.Error("Failed to validate create request", slog.Any("err", err))
		return nil, err
	}

	slog.Info("Creating subscription", slog.String("source", req.Source), slog.String("eventType", req.EventType))

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
		slog.Error("Failed to create subscription", slog.Any("err", err))
		return nil, err
	}

	slog.Info("Subscription created", slog.String("id", sub.Id))

	return sub, nil
}

func (s Service) GetByID(ctx context.Context, id string) (*subscription.Subscription, error) {
	slog.Info("Getting subscription by ID", slog.String("id", id))

	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		slog.Error("Failed to get subscription by ID", slog.String("id", id), slog.Any("err", err))
		return nil, err
	}

	return sub, nil
}

func (s Service) List(ctx context.Context, source, eventType string) ([]subscription.Subscription, error) {
	slog.Info("Listing subscriptions", slog.String("source", source), slog.String("eventType", eventType))

	subs, err := s.repo.List(ctx, source, eventType)
	if err != nil {
		slog.Error("Failed to list subscriptions", slog.Any("err", err))
		return nil, err
	}

	slog.Info("Listed subscriptions", slog.Int("count", len(subs)))

	return subs, nil
}

func (s Service) Update(ctx context.Context, id string, req *subscription.UpdateRequest) (*subscription.Subscription, error) {
	if err := req.Validate(); err != nil {
		slog.Error("Failed to validate update request", slog.String("id", id), slog.Any("err", err))
		return nil, err
	}

	slog.Info("Updating subscription", slog.String("id", id))

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		slog.Error("Failed to get subscription for update", slog.String("id", id), slog.Any("err", err))
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
		slog.Error("Failed to update subscription", slog.String("id", id), slog.Any("err", err))
		return nil, err
	}

	slog.Info("Subscription updated", slog.String("id", id))

	return existing, nil
}

func (s Service) Delete(ctx context.Context, id string) error {
	slog.Info("Deleting subscription", slog.String("id", id))

	if err := s.repo.Delete(ctx, id); err != nil {
		slog.Error("Failed to delete subscription", slog.String("id", id), slog.Any("err", err))
		return err
	}

	slog.Info("Subscription deleted", slog.String("id", id))

	return nil
}
