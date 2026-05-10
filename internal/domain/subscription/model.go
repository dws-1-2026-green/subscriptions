package subscription

import "time"

type Subscription struct {
	Id             string
	Source         string
	EventType      string
	DestinationUrl string
	Method         string
	Headers        map[string]string
	Enabled        bool
	CreatedAt      time.Time
}
