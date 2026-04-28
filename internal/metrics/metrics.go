package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	EventsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "subscriptions_events_processed_total",
		Help: "Total number of routing.requests events processed",
	}, []string{"source", "event_type"})

	DeliveriesDispatched = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "subscriptions_deliveries_dispatched_total",
		Help: "Total number of delivery tasks sent to Kafka",
	}, []string{"source", "event_type"})

	NoMatchesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "subscriptions_no_matches_total",
		Help: "Events that matched zero active subscriptions",
	}, []string{"source", "event_type"})

	RoutingErrorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "subscriptions_routing_errors_total",
		Help: "Total DB or routing errors (message not committed)",
	})

	DBQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "subscriptions_db_query_duration_seconds",
		Help:    "Latency of subscription lookup queries",
		Buckets: prometheus.DefBuckets,
	}, []string{"backend"})
)
