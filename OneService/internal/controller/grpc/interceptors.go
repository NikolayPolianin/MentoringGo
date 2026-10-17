package grpc

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"log/slog"
)

func UnaryServerInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// Логирование перед вызовом
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		auth := md["authorization"]
		if len(auth) == 0 {
			return nil, status.Error(codes.Unauthenticated, "no auth")
		}
		slog.Info("UnaryServerInterceptor", "auth", auth)
	}
	// Вызов основного обработчика
	resp, err := handler(ctx, req)
	// Логирование после вызова
	log.Printf("After handling: %v", info.FullMethod)
	return resp, err
}
