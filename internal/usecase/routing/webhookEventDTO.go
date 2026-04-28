package routing

import "encoding/json"

// WebhookEventDTO represents the event data included in a webhook delivery payload.
type WebhookEventDTO struct {
	Id   string          `json:"id"`
	Data json.RawMessage `json:"data"`
}