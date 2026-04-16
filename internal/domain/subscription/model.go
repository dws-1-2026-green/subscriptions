package subscription

type Subscription struct {
	Id              string
	DestinnationUrl string
	Method          string
	Headers         map[string]string
}
