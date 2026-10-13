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
	"checkgo/proto"
	"context"
	httpSwagger "github.com/swaggo/http-swagger"
	grpcServer "google.golang.org/grpc"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"time"
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

	producer, err := kafka.NewProducer()
	if err != nil {
		slog.Error("Failed to create producer", err.Error())
	}

	outboxCfg := worker.NewOutboxConfig(3)

	useCase := use_case.NewUseCase(pg, producer, cache)
	hu := handler.NewUserHandler(useCase)
	userRouter := http2.NewRouter(hu)

	userRouter.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))
	outboxRunner := worker.NewOutboxKafka(useCase, *outboxCfg)

	server := http_server.StartHttpServer(userRouter)

	if err = outboxRunner.Run(ctx); err != nil {
		slog.Error("Failed to start outbox", err.Error())
	}

	lst, err := net.Listen("tcp", ":50051")
	if err != nil {
		slog.Error("Failed to create listener", err.Error())
	}

	s := grpcServer.NewServer(
		grpcServer.UnaryInterceptor(grpc2.UnaryServerInterceptor))
	proto.RegisterUserServiceServer(s, &grpc2.Server{UseCase: useCase})
	if err := s.Serve(lst); err != nil {
		slog.Error("Failed to start server", err.Error())
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch

	outboxRunner.Stop()
	server.Stop()
	pg.Close()
	cache.Close()
	producer.Close()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		<-ctx.Done()
		if err := lst.Close(); err != nil {
			slog.Error("Failed to close listener", err.Error())
		}
	}()

}
