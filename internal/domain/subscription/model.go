package subscription

type Subscription struct {
	Id             string
	DestinationUrl string
	Method         string
	Headers        map[string]string
}