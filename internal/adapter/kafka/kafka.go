package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/dws-1-2026-green/subscriptions/internal/metrics"
	"github.com/dws-1-2026-green/subscriptions/internal/usecase/routing"
	"github.com/dws-1-2026-green/subscriptions/internal/worker"
	kafkago "github.com/segmentio/kafka-go"
)

type KafkaWorker struct {
	reader      *kafkago.Reader
	writer      *kafkago.Writer
	handler     routing.Handler
	concurrency int
}

// procToken carries a fetched message through the parallel pipeline.
// The committer goroutine processes tokens in FIFO order; worker goroutines
// fill them concurrently. This preserves at-least-once semantics: offsets are
// committed only after successful processing, and always in ascending order.
type procToken struct {
	msg  kafkago.Message
	out  []kafkago.Message // deliveries to publish before committing
	err  error
	done chan struct{} // closed by the worker goroutine when finished
}

func (kw *KafkaWorker) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Buffer = concurrency: limits in-flight messages and provides back-pressure.
	tokens := make(chan *procToken, kw.concurrency)
	commitErr := make(chan error, 1)

	// Committer: reads tokens in FIFO order, ensuring offsets always advance.
	go func() {
		defer close(commitErr)
		for t := range tokens {
			select {
			case <-t.done:
			case <-ctx.Done():
				return
			}

			if t.err != nil {
				slog.Error("worker: processing failed, not committing", slog.Any("err", t.err))
				// Cancel causes the fetcher to stop; unprocessed messages will
				// be redelivered on the next consumer startup.
				cancel()
				return
			}

			if len(t.out) > 0 {
				if err := kw.writer.WriteMessages(ctx, t.out...); err != nil {
					slog.Error("worker: failed to write deliveries", slog.Any("err", err))
					cancel()
					return
				}
			}

			if err := kw.reader.CommitMessages(ctx, t.msg); err != nil {
				if ctx.Err() == nil {
					commitErr <- fmt.Errorf("commit message: %w", err)
				}
				return
			}

			slog.Info("committed", slog.Int64("offset", t.msg.Offset))
		}
	}()

	for {
		msg, err := kw.reader.FetchMessage(ctx)
		if err != nil {
			close(tokens)
			if cerr := <-commitErr; cerr != nil {
				return cerr
			}
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("fetch message: %w", err)
		}

		slog.Info("received routing request",
			slog.String("key", string(msg.Key)),
			slog.Int64("offset", msg.Offset),
		)

		var event routing.RoutingRequestDTO
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			slog.Error("unmarshal error, skipping message",
				slog.Int64("offset", msg.Offset),
				slog.Any("err", err),
			)
			_ = kw.reader.CommitMessages(ctx, msg)
			continue
		}

		metrics.KafkaMessagesConsumed.Inc()

		t := &procToken{msg: msg, done: make(chan struct{})}

		select {
		case tokens <- t:
		case <-ctx.Done():
			close(tokens)
			<-commitErr
			return nil
		}

		go func(t *procToken, evt routing.RoutingRequestDTO) {
			defer close(t.done)
			t.out, t.err = kw.process(ctx, evt)
		}(t, event)
	}
}

func (kw *KafkaWorker) process(ctx context.Context, event routing.RoutingRequestDTO) ([]kafkago.Message, error) {
	webhooks, err := kw.handler.GetDestinationUrl(ctx, event)
	if err != nil {
		return nil, err
	}

	if len(webhooks) == 0 {
		return nil, nil
	}

	out := make([]kafkago.Message, 0, len(webhooks))
	for _, wh := range webhooks {
		b, err := json.Marshal(wh)
		if err != nil {
			return nil, fmt.Errorf("marshal webhook: %w", err)
		}
		out = append(out, kafkago.Message{
			Key:   []byte(wh.DeliveryId),
			Value: b,
		})
	}
	return out, nil
}

func NewWorker(reader *kafkago.Reader, writer *kafkago.Writer, handler routing.Handler, concurrency int) worker.Worker {
	return &KafkaWorker{
		reader:      reader,
		writer:      writer,
		handler:     handler,
		concurrency: concurrency,
	}
}
