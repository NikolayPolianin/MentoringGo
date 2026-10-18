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
	"checkgo/observability/metrics"
	"checkgo/observability/otel"
	"context"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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
	trace := otel.InitTracer(ctx)
	if trace != nil {
		slog.Error("OTel tracer initialized", "err", trace)
	}
	metric := metrics.InitMetrics(ctx)
	if metric != nil {
		slog.Info("OTel metrics initialized", "err", metric)
	}

	pg := postgres.NewPostgres(ctx)
	pg.CreateTable(ctx)
	cache := redis.NewRedisClient()
	tx := postgres.NewTxManager(pg.Pool)
	// Metrics
	httpMetrics := metrics.NewHTTPServer()

	producer, err := kafka.NewProducer()
	if err != nil {
		slog.Error("Failed to create producer", err.Error())
	}

	outboxCfg := worker.NewOutboxConfig(10)

	useCase := use_case.NewUseCase(pg, producer, cache, tx)
	hu := handler.NewUserHandler(useCase)
	router := http2.NewRouter(hu, httpMetrics)
	outboxRunner := worker.NewOutboxKafka(useCase, *outboxCfg)

	handler := otelhttp.NewHandler(
		router,
		"user-service",
	)
	httpServer := http_server.StartHttpServer(handler)
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
