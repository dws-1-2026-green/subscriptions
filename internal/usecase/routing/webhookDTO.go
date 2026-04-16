package routing

import "encoding/json"

type WebhookDTO struct {
	DeliveryId string

	Event struct {
		Id   string
		Data json.RawMessage
	}

	Subscription struct {
		Id             string
		DestinationUrl string
		Method         string
		Headers        map[string]string
	}

	MappedAt string
	TraceId  string
}
