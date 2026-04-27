package routing

import "encoding/json"

type WebhookDTO struct {
	DeliveryId string `json:"delivery_id"`

	Event struct {
		Id   string          `json:"id"`
		Data json.RawMessage `json:"data"`
	} `json:"event"`

	Subscription struct {
		Id             string            `json:"id"`
		DestinationUrl string            `json:"destination_url"`
		Method         string            `json:"method"`
		Headers        map[string]string `json:"headers"`
	} `json:"subscription"`

	MappedAt string `json:"mapped_at"`
	TraceId  string `json:"trace_id"`
}