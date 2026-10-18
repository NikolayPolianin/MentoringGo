package http

import (
	"checkgo/OneService/internal/transport/http/handler"
	"checkgo/observability/metrics"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
)

type Router struct {
}

func NewRouter(userHandler *handler.UserHandler, m *metrics.HTTPServer) http.Handler {
	r := chi.NewRouter()

	r.Use(metrics.NewMiddleware(m))
	r.Use(middleware.Recoverer)

	r.Route("/api/v1/", func(r chi.Router) {
		r.Post("/jwt/{id}", userHandler.CreateAuthToken)
		r.Route("/users", func(r chi.Router) {
			r.Post("/", userHandler.CreateUser)
			r.Get("/{id}", userHandler.GetUser)
		})
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	return r
}
