package event

import "encoding/json"

type Event struct {
	Id     string
	Source string
	Type   string
	Data   json.RawMessage
}
