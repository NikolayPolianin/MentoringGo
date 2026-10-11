package app

import (
	"checkgo/TwoService/internal/messaging/kafka"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func Run(ctx context.Context) {
	consumer := kafka.NewConsumer()
	go func() {
		consumer.Run(ctx)
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	<-sig // wait signal

	slog.Info("Get Graceful Shutdown command")
	consumer.Close()
}
