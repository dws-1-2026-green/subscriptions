package routing

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Handler handles routing of webhook requests to subscriptions.
type Handler struct {
	repo SubscriptionsRepo
}

// GetDestinationUrl retrieves webhook destinations for a given event.
func (h Handler) GetDestinationUrl(ctx context.Context, event RoutingRequestDTO) ([]WebhookDTO, error) {
	subscriptions, err := h.repo.ListBySourceAndType(ctx, event.Event.Source, event.Event.Type)

	if err != nil {
		return []WebhookDTO{}, err
	}

	webhooks := make([]WebhookDTO, 0, len(subscriptions))

	for _, sub := range subscriptions {
		webhook := WebhookDTO{
			DeliveryId: uuid.NewString(),
			Event: WebhookEventDTO{
				Id:   event.Event.Id,
				Data: event.Event.Data,
			},
			Subscription: WebhookSubscriptionDTO{
				Id:             sub.Id,
				DestinationUrl: sub.DestinationUrl,
				Method:         sub.Method,
				Headers:        sub.Headers,
			},
			MappedAt: time.Now().Format(time.RFC3339),
			TraceId:  event.TraceId,
		}

		webhooks = append(webhooks, webhook)
	}

	return webhooks, nil
}

// NewHandler creates a new Handler with the given subscriptions repository.
func NewHandler(repo SubscriptionsRepo) Handler {
	return Handler{
		repo: repo,
	}
}