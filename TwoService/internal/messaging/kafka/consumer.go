package kafka

import (
	"checkgo/TwoService/internal/domain"
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"log/slog"
	"time"
)

type Consumer struct {
	group sarama.ConsumerGroup
}
type ConsumerHandler struct{}

func NewConsumer() *Consumer {
	config := sarama.NewConfig()
	config.Consumer.Retry.Max = 3
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.AutoCommit.Enable = false

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
		var user domain.User
		if err := json.Unmarshal(message.Value, &user); err != nil {
			slog.Error("Error unmarshalling message:", err)
		}
		if err := json.Unmarshal(message.Key, &user.UuID); err != nil {
			slog.Error("Error unmarshalling Key:", err)
		}
		slog.Info("Message claimed:",
			"Topic:", message.Topic,
			"User:", user)
	}
	return nil
}
func (c *Consumer) Run(ctx context.Context) error {
	handler := &ConsumerHandler{}
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err := c.group.Consume(ctx, []string{"user-create", "user-skip"}, handler)
		if err != nil {
			slog.Error("Error from consumer:", err)
			return err
		}
		slog.Info("Consumer successfully initiated")
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
