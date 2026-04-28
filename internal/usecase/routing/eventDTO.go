package routing

import "encoding/json"

type RoutingRequestEventDTO struct {
	Id     string `json:"id"`
	Source string `json:"source"`
	Type   string `json:"type"`

	Data json.RawMessage `json:"data"`
}

type RoutingRequestDTO struct {
	Event RoutingRequestEventDTO `json:"event"`

	IngestedAt string `json:"ingested_at"`
	TraceId    string `json:"trace_id"`
}
