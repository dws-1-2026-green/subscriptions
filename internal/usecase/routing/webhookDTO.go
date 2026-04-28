package routing

// WebhookDTO represents the delivery payload published to the deliveries.to_send Kafka topic.
type WebhookDTO struct {
	DeliveryId   string                 `json:"delivery_id"`
	Event        WebhookEventDTO        `json:"event"`
	Subscription WebhookSubscriptionDTO `json:"subscription"`
	MappedAt     string                 `json:"mapped_at"`
	TraceId      string                 `json:"trace_id"`
}