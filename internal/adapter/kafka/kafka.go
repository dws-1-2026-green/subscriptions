package kafka

import (
	"context"

	"github.com/dws-1-2026-green/subscriptions/internal/usecase/routing"
	"github.com/dws-1-2026-green/subscriptions/internal/worker"
	kafkago "github.com/segmentio/kafka-go"
)

type KafkaWorker struct {
	reader  *kafkago.Reader
	writer  *kafkago.Writer
	handler routing.Handler
}

func (kw KafkaWorker) Run(ctx context.Context) error {
	return nil
}

func NewWorker(reader *kafkago.Reader, writer *kafkago.Writer, handler routing.Handler) worker.Worker {
	return KafkaWorker{
		reader:  reader,
		writer:  writer,
		handler: handler,
	}
}
