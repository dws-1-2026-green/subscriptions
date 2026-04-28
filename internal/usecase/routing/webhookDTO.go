package routing

import "encoding/json"

type WebhookEventDTO struct {
	Id   string          `json:"id"`
	Data json.RawMessage `json:"data"`
}

type WebhookSubscriptionDTO struct {
	Id             string            `json:"id"`
	DestinationUrl string            `json:"destination_url"`
	Method         string            `json:"method"`
	Headers        map[string]string `json:"headers"`
}

type WebhookDTO struct {
	DeliveryId string `json:"delivery_id"`

	Event WebhookEventDTO `json:"event"`

	Subscription WebhookSubscriptionDTO `json:"subscription"`

	MappedAt string `json:"mapped_at"`
	TraceId  string `json:"trace_id"`
}
