package http

import (
	"checkgo/OneService/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
)

type Router struct {
	uh *handler.UserHandler
}

func NewRouter(handler *handler.UserHandler) *chi.Mux {
	r := chi.NewRouter()
	router := Router{uh: handler}

	r.Post("/jwt/{id}", router.uh.CreateAuthToken)
	r.Route("/users", func(r chi.Router) {
		r.Post("/", router.uh.CreateUser)
		r.Get("/{id}", router.uh.GetUser)
	})

	return r
}
