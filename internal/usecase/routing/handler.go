package routing

type Handler struct {
	repo SubscriptionsRepo
}

func (h Handler) GetDestinationUrl(event EventDTO) WebhookDTO {
	return WebhookDTO{}
}

func NewHandler(repo SubscriptionsRepo) Handler {
	return Handler{
		repo: repo,
	}
}
