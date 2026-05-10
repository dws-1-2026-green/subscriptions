package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"

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
	for {
		msg, err := kw.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("fetch message: %w", err)
		}

		slog.Info("Received routing request", slog.String("key", string(msg.Key)), slog.Int64("offset", msg.Offset))

		var event routing.RoutingRequestDTO
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			return fmt.Errorf("unmarshal RoutingRequestDTO: %w", err)
		}

		webhooks, err := kw.handler.GetDestinationUrl(ctx, event)
		if err != nil {
			slog.Error("Failed to get destination URLs", slog.Any("err", err))

			// do not commit -> Kafka will re-read the message
			continue
		}

		log.Printf("matched %d webhooks for event source=%s type=%s", len(webhooks), event.Event.Source, event.Event.Type)

		if len(webhooks) > 0 {
			out := make([]kafkago.Message, 0, len(webhooks))

			for _, wh := range webhooks {
				b, err := json.Marshal(wh)
				if err != nil {
					// не коммитим -> перечитаем
					slog.Info("Failed to marshal delivery", slog.Any("err", err))
					out = nil
					break
				}

				// Ключ лучше делать стабильным для идемпотентности/упорядочивания.
				// DeliveryId подходит лучше всего.
				out = append(out, kafkago.Message{
					Key:   []byte(wh.DeliveryId),
					Value: b,
				})
			}

			if out == nil {
				continue
			}

			if err := kw.writer.WriteMessages(ctx, out...); err != nil {
				// не коммитим -> перечитаем
				slog.Error("Failed to write delivery messages", slog.Any("err", err))
				continue
			}
		}

		if err := kw.reader.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("commit message: %w", err)
		}

		slog.Info("Processed", slog.Int64("offset", msg.Offset))
	}
}

func NewWorker(reader *kafkago.Reader, writer *kafkago.Writer, handler routing.Handler) worker.Worker {
	return KafkaWorker{
		reader:  reader,
		writer:  writer,
		handler: handler,
	}
}
