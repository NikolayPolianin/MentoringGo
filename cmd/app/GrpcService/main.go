package main

import (
	"checkgo/proto"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"time"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	ticker := time.NewTicker(5 * time.Second)
	user := &proto.CreateUserRequest{
		Name:     "TestGRPC",
		Age:      23,
		DateBorn: "2000-05-11",
	}
	md := metadata.New(map[string]string{
		"authorization": "Bearer your_token",
		"user-id":       "123456",
	})
	c := proto.NewUserServiceClient(conn)
	for {
		select {
		case <-ticker.C:
			ctx := metadata.NewOutgoingContext(context.Background(), md)
			r, err := c.CreateUser(ctx, user)
			if err != nil {
				log.Fatalf("could not create user: %v", err)
			}
			slog.Info("created user:", "Id:", r.Id)
		}
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	conn.Close()
}
