package config

import "github.com/caarlos0/env/v11"

type Config struct {
	StoreBackend string `env:"STORE_BACKEND" envDefault:"postgres"` // postgres|cassandra

	// postgres
	DatabaseURL string `env:"DATABASE_URL"`

	// cassandra
	CassandraHosts       []string `env:"CASSANDRA_HOSTS" envSeparator:","`
	CassandraKeyspace    string   `env:"CASSANDRA_KEYSPACE" envDefault:"webhooks"`
	CassandraConsistency string   `env:"CASSANDRA_CONSISTENCY" envDefault:"QUORUM"`

	// kafka
	KafkaBrokers          []string `env:"KAFKA_BROKERS,required" envSeparator:","`
	KafkaGroupID          string   `env:"KAFKA_GROUP_ID" envDefault:"subscriptions-worker"`
	RoutingRequestsTopic  string   `env:"KAFKA_ROUTING_REQUESTS_TOPIC" envDefault:"routing.requests"`
	DeliveriesToSendTopic string   `env:"KAFKA_DELIVERIES_TOPIC" envDefault:"deliveries.to_send"`
}

func Load() (Config, error) {
	var c Config
	return c, env.Parse(&c)
}
