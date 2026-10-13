package kafka

import (
	"checkgo/OneService/internal/domain"
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

func (p *Producer) SendMessage(ctx context.Context, events ...*domain.Event) error {
	for _, event := range events {

		if event == nil {
			continue
		}
		msg := sarama.ProducerMessage{
			Topic: event.Topic,
			Key:   sarama.StringEncoder(event.Key),
			Value: sarama.StringEncoder(event.Value),
		}
		part, off, _ := p.producer.SendMessage(&msg)
		slog.Info("IN:", part, off)
	}
	return nil
}
