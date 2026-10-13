package consumer

import (
	"checkgo/TwoService/internal/use_case"
	"context"
	"github.com/IBM/sarama"
	"log/slog"
	"time"
)

type Consumer struct {
	group sarama.ConsumerGroup
}
type ConsumerHandler struct {
	uc *use_case.InboxUseCase
}

func NewConsumer() *Consumer {
	config := sarama.NewConfig()
	config.Consumer.Retry.Max = 3
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumerGroup([]string{"localhost:9092", "localhost:9093"}, "user-service", config)
	if err != nil {
		slog.Error("Error creating consumer:", err)
	}

	return &Consumer{group: consumer}
}
func (h *ConsumerHandler) Setup(session sarama.ConsumerGroupSession) error {
	slog.Info("Consumer setup called:", session.Claims())
	return nil
}
func (h *ConsumerHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}
func (h *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		if err := h.uc.ReadAndWriteInbox(context.Background(), message); err != nil {
			session.MarkMessage(message, "")
			return err
		}
		session.MarkMessage(message, "")
		slog.Info(
			"processing Kafka message",
			"topic", message.Topic,
			"partition", message.Partition,
			"offset", message.Offset,
			"key", string(message.Key),
		)
	}
	return nil
}
func (c *Consumer) Run(ctx context.Context, uc *use_case.InboxUseCase) error {
	handler := &ConsumerHandler{uc}
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err := c.group.Consume(ctx, []string{"user-create", "user-skip"}, handler)
		if err != nil {
			slog.Error("Error from consumer:", "err", err)
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second * 10):
		}
	}
}
func (c *Consumer) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		if err := c.group.Close(); err != nil {
			slog.Error("Error closing consumer:", err)
		}
	}
}
