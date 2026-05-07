package subscription

import "errors"

var (
	ErrSourceRequired         = errors.New("source is required")
	ErrEventTypeRequired      = errors.New("event_type is required")
	ErrDestinationURLRequired = errors.New("destination_url is required")
	ErrNothingToUpdate        = errors.New("no fields to update")
	ErrNotFound               = errors.New("subscription not found")
)
