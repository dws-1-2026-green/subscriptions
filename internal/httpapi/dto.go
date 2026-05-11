package httpapi

import "time"

type SubscriptionDTO struct {
	Id             string            `json:"subscription_id"`
	Source         string            `json:"source"`
	EventType      string            `json:"event_type"`
	DestinationUrl string            `json:"destination_url"`
	Method         string            `json:"http_method"`
	Headers        map[string]string `json:"headers,omitempty"`
	Enabled        bool              `json:"enabled"`
	CreatedAt      time.Time         `json:"created_at"`
}
