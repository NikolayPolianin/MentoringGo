package kafka

import (
	"context"
	"github.com/IBM/sarama"
	"log/slog"
	"time"
)

type Producer struct {
	producer sarama.SyncProducer
}

func NewProducer() (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll

	producer, err := sarama.NewSyncProducer([]string{"localhost:9092", "localhost:9093"}, config)
	if err != nil {
		return nil, err
	}
	return &Producer{producer}, nil
}

func (p *Producer) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	<-ctx.Done()
	if err := p.producer.Close(); err != nil {
		slog.Error("Failed to close producer")
	}
}

func (p *Producer) SendMessage(message *sarama.ProducerMessage) error {
	partition, offset, _ := p.producer.SendMessage(message)
	value, err := message.Value.Encode()
	if err != nil {
		slog.Error("failed to encode message value", "err", err)
		return err
	}
	slog.Info("message sent",
		"topic", message.Topic,
		"key", message.Key,
		"value", string(value),
		"in partition: ", partition,
		"WithOffset", offset)
	return nil
}
