package router

import (
	"github.com/fc1g/kachenyatko-store-api/internal/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const maxBodyBytes = 1 << 20

func New(categoryHandler *handler.CategoryHandler, toyHandler *handler.ToyHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestSize(maxBodyBytes))

	r.Route("/categories", func(r chi.Router) {
		r.Post("/", categoryHandler.Create)
		r.Get("/", categoryHandler.GetAll)
	})

	r.Route("/toys", func(r chi.Router) {
		r.Post("/", toyHandler.Create)
		r.Get("/", toyHandler.GetAll)
	})

	return r
}
