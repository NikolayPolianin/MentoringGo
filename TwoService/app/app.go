package app

import (
	"checkgo/TwoService/internal/adapter/postgres"
	"checkgo/TwoService/internal/controller/consumer"
	"checkgo/TwoService/internal/use_case"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func Run(ctx context.Context) {
	pg := postgres.NewPostgres(ctx)
	uc := use_case.NewInboxUseCase(pg)
	consumer := consumer.NewConsumer()

	go func() {
		consumer.Run(ctx, uc)
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	<-sig // wait signal

	slog.Info("Get Graceful Shutdown command")
	consumer.Close()
}
