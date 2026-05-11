package subscription

type CreateRequest struct {
	Source         string            `json:"source"`
	EventType      string            `json:"event_type"`
	DestinationUrl string            `json:"destination_url"`
	Method         string            `json:"http_method"`
	Headers        map[string]string `json:"headers,omitempty"`
}

// Validate validates the create request
func (r *CreateRequest) Validate() error {
	if r.Source == "" {
		return ErrSourceRequired
	}
	if r.EventType == "" {
		return ErrEventTypeRequired
	}
	if r.DestinationUrl == "" {
		return ErrDestinationURLRequired
	}
	if r.Method == "" {
		r.Method = "POST"
	}
	if r.Headers == nil {
		r.Headers = make(map[string]string)
	}
	return nil
}

type UpdateRequest struct {
	DestinationUrl string            `json:"destination_url"`
	Method         string            `json:"http_method"`
	Headers        map[string]string `json:"headers,omitempty"`
	Enabled        *bool             `json:"enabled,omitempty"`
}

func (r *UpdateRequest) Validate() error {
	if r.DestinationUrl == "" && r.Method == "" && r.Headers == nil && r.Enabled == nil {
		return ErrNothingToUpdate
	}
	if r.Headers == nil {
		r.Headers = make(map[string]string)
	}
	return nil
}
