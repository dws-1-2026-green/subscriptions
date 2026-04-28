package routing

import "encoding/json"

// RoutingRequestEventDTO represents the event data in a routing request.
type RoutingRequestEventDTO struct {
	Id     string `json:"id"`
	Source string `json:"source"`
	Type   string `json:"type"`

	Data json.RawMessage `json:"data"`
}

// RoutingRequestDTO represents the complete routing request payload received from Kafka.
type RoutingRequestDTO struct {
	Event      RoutingRequestEventDTO `json:"event"`
	IngestedAt string                 `json:"ingested_at"`
	TraceId    string                 `json:"trace_id"`
}