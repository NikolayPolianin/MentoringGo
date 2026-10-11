package http

import (
	"checkgo/OneService/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
)

type Router struct {
	uh *handler.UserHandler
}

func NewRouter(handler *handler.UserHandler) *Router {
	return &Router{uh: handler}
}
func (router *Router) RegisterRoutes(r chi.Router) {
	r.Post("/jwt/{id}", router.uh.CreateAuthToken)
	r.Route("/users", func(r chi.Router) {
		r.Post("/", router.uh.CreateUser)
		r.Get("/{id}", router.uh.GetUser)
	})
}
