package http_server

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type HttpServer struct {
	httpServer *http.Server
}

func StartHttpServer(handler http.Handler) *HttpServer {
	client := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	s := &HttpServer{
		httpServer: client,
	}

	go s.start()
	slog.Info("http server start")

	return s
}

func (s *HttpServer) start() {
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("http server start error", "error:", err.Error())
	}
}
func (s *HttpServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	<-ctx.Done()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		slog.Error("http server shutdown error", "error:", err.Error())
	}
}
