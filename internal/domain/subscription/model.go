package subscription

// Subscription represents a webhook subscription configuration.
type Subscription struct {
	Id             string
	DestinationUrl string
	Method         string
	Headers        map[string]string
}