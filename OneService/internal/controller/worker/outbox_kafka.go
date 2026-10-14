package worker

import (
	"checkgo/OneService/internal/use_case"
	"context"
	"log/slog"
	"time"
)

type OutboxConfig struct {
	BatchSize int `env:"BATCH_SIZE" envDefault:"10"`
}
type OutboxKafka struct {
	uc           *use_case.User
	OutboxConfig OutboxConfig
	stop         chan struct{}
	done         chan struct{}
}

func NewOutboxConfig(BatchSize int) *OutboxConfig {
	return &OutboxConfig{
		BatchSize: BatchSize,
	}
}
func NewOutboxKafka(uc *use_case.User, outboxConfig OutboxConfig) *OutboxKafka {
	return &OutboxKafka{
		uc:           uc,
		OutboxConfig: outboxConfig,
		stop:         make(chan struct{}),
		done:         make(chan struct{}),
	}
}

func (o *OutboxKafka) Run(ctx context.Context) error {
	go func() {
	FOR:
		for {
			events, err := o.uc.OutboxReadAndProduce(ctx, o.OutboxConfig.BatchSize)
			if err != nil {
				slog.Error("sd", "err", err)
			}
			var duration time.Duration
			if events < o.OutboxConfig.BatchSize {
				slog.Info("events is too small - ", "count", events)
				duration = 10 * time.Second
			}
			select {
			case <-o.stop:
				break FOR
			case <-time.After(duration):
			}
		}
	}()
	return nil
}
func (o *OutboxKafka) Stop() {
	close(o.stop)
	<-o.done
}
