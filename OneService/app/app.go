package app

import (
	_ "checkgo/OneService/docs"
	"checkgo/OneService/internal/adapter/kafka"
	"checkgo/OneService/internal/adapter/postgres"
	"checkgo/OneService/internal/adapter/redis"
	grpc2 "checkgo/OneService/internal/controller/grpc"
	"checkgo/OneService/internal/controller/worker"
	"checkgo/OneService/internal/http_server"
	http2 "checkgo/OneService/internal/transport/http"
	"checkgo/OneService/internal/transport/http/handler"
	"checkgo/OneService/internal/use_case"
	"context"
	httpSwagger "github.com/swaggo/http-swagger"
	"log/slog"
	"os"
	"os/signal"
)

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @title CheckGo API
// @version 1.0
// @description API for CheckGo
// @host localhost:8080
// @BasePath /
func Run(ctx context.Context) {
	pg := postgres.NewPostgres(ctx)
	pg.CreateTable(ctx)
	cache := redis.NewRedisClient()
	tx := postgres.NewTxManager(pg.Pool)

	producer, err := kafka.NewProducer()
	if err != nil {
		slog.Error("Failed to create producer", err.Error())
	}

	outboxCfg := worker.NewOutboxConfig(10)

	useCase := use_case.NewUseCase(pg, producer, cache, tx)
	hu := handler.NewUserHandler(useCase)
	userRouter := http2.NewRouter(hu)
	outboxRunner := worker.NewOutboxKafka(useCase, *outboxCfg)

	userRouter.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	httpServer := http_server.StartHttpServer(userRouter)
	go grpc2.StartServer(ctx, useCase)

	if err = outboxRunner.Run(ctx); err != nil {
		slog.Error("Failed to start outbox", err.Error())
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch

	outboxRunner.Stop()
	httpServer.Stop()
	pg.Close()
	cache.Close()
	producer.Close()
}
