package routing

import "encoding/json"

type EventDTO struct {
	Event struct {
		Id     string
		Source string
		Type   string

		Data json.RawMessage
	}

	IngestedAt string
	TraceId    string
}
