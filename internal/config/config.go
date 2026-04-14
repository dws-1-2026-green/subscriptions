package config

import "github.com/caarlos0/env/v11"

type Config struct {
	StoreBackend            string `env:"STORE_BACKEND" envDefault:"postgres"` // postgres|http
	DatabaseURL             string `env:"DATABASE_URL"`
	SubscriptionsAPIBaseURL string `env:"SUBSCRIPTIONS_API_BASE_URL"`

	KafkaBrokers          []string `env:"KAFKA_BROKERS,required" envSeparator:","`
	KafkaGroupID          string   `env:"KAFKA_GROUP_ID" envDefault:"subscriptions-worker"`
	RoutingRequestsTopic  string   `env:"KAFKA_ROUTING_REQUESTS_TOPIC" envDefault:"routing.requests"`
	DeliveriesToSendTopic string   `env:"KAFKA_DELIVERIES_TOPIC" envDefault:"deliveries.to_send"`
}

func Load() (Config, error) {
	var c Config
	return c, env.Parse(&c)
}
