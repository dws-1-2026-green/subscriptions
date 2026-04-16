package routing

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Handler struct {
	repo SubscriptionsRepo
}

func (h Handler) GetDestinationUrl(ctx context.Context, event EventDTO) ([]WebhookDTO, error) {
	subscriptions, err := h.repo.ListBySourceAndType(ctx, event.Event.Source, event.Event.Type)

	if err != nil {
		return []WebhookDTO{}, err
	}

	webhooks := make([]WebhookDTO, 0, len(subscriptions))

	for _, sub := range subscriptions {
		webhook := WebhookDTO{
			DeliveryId: uuid.NewString(),

			Event: struct {
				Id   string
				Data json.RawMessage
			}{
				Id:   event.Event.Id,
				Data: event.Event.Data,
			},

			Subscription: struct {
				Id             string
				DestinationUrl string
				Method         string
				Headers        map[string]string
			}{
				Id:             sub.Id,
				DestinationUrl: sub.DestinnationUrl,
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

func NewHandler(repo SubscriptionsRepo) Handler {
	return Handler{
		repo: repo,
	}
}
