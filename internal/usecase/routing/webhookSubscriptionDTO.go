package routing

// WebhookSubscriptionDTO represents the subscription configuration in a webhook delivery payload.
type WebhookSubscriptionDTO struct {
	Id             string            `json:"id"`
	DestinationUrl string            `json:"destination_url"`
	Method         string            `json:"method"`
	Headers        map[string]string `json:"headers"`
}