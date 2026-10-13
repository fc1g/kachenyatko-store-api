package router

import (
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/httpmw"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

const maxBodyBytes = 1 << 20

type healthHandler interface {
	Live(w http.ResponseWriter, r *http.Request)
	Ready(w http.ResponseWriter, r *http.Request)
}

type categoryHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	GetAll(w http.ResponseWriter, r *http.Request)
	GetOne(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

type toyHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	GetAll(w http.ResponseWriter, r *http.Request)
	GetOne(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

type Handlers struct {
	Health   healthHandler
	Category categoryHandler
	Toy      toyHandler
}

func New(logger zerolog.Logger, handlers Handlers) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(httpmw.RequestLogger(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestSize(maxBodyBytes))

	r.Get("/livez", handlers.Health.Live)
	r.Get("/readyz", handlers.Health.Ready)

	r.Route("/categories", func(r chi.Router) {
		r.Post("/", handlers.Category.Create)
		r.Get("/", handlers.Category.GetAll)
		r.Get("/{categoryID}", handlers.Category.GetOne)
		r.Put("/{categoryID}", handlers.Category.Update)
		r.Delete("/{categoryID}", handlers.Category.Delete)
	})

	r.Route("/toys", func(r chi.Router) {
		r.Post("/", handlers.Toy.Create)
		r.Get("/", handlers.Toy.GetAll)
		r.Get("/{toyID}", handlers.Toy.GetOne)
		r.Put("/{toyID}", handlers.Toy.Update)
		r.Delete("/{toyID}", handlers.Toy.Delete)
	})

	return r
}
