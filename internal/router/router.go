package router

import (
	"net/http"

	_ "github.com/fc1g/kachenyatko-store-api/docs"
	"github.com/fc1g/kachenyatko-store-api/internal/httpmw"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

const (
	maxBodyBytes        = 1 << 20
	maxImageUploadBytes = 51 << 20
)

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

type toyImageHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	GetAll(w http.ResponseWriter, r *http.Request)
	GetOne(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

type Handlers struct {
	Health   healthHandler
	Category categoryHandler
	Toy      toyHandler
	ToyImage toyImageHandler
}

func New(logger zerolog.Logger, handlers Handlers) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(httpmw.RequestLogger(logger))
	r.Use(middleware.Recoverer)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Get("/livez", handlers.Health.Live)
	r.Get("/readyz", handlers.Health.Ready)

	r.Route("/categories", func(r chi.Router) {
		r.Get("/", handlers.Category.GetAll)
		r.Get("/{slug}", handlers.Category.GetOne)

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequestSize(maxBodyBytes))
			// TODO: r.use(requireAdmin)
			r.Post("/", handlers.Category.Create)
			r.Put("/{categoryID}", handlers.Category.Update)
			r.Delete("/{categoryID}", handlers.Category.Delete)
		})
	})

	r.Route("/toys", func(r chi.Router) {
		r.Get("/", handlers.Toy.GetAll)
		r.Get("/{toyID}", handlers.Toy.GetOne)

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequestSize(maxBodyBytes))
			// TODO: r.use(requireAdmin)
			r.Post("/", handlers.Toy.Create)
			r.Put("/{toyID}", handlers.Toy.Update)
			r.Delete("/{toyID}", handlers.Toy.Delete)
		})

		r.Route("/{toyID}/images", func(r chi.Router) {
			r.Get("/", handlers.ToyImage.GetAll)
			r.Get("/{imageID}", handlers.ToyImage.GetOne)

			r.Group(func(r chi.Router) {
				// TODO: r.use(requireAdmin)
				r.With(middleware.RequestSize(maxImageUploadBytes)).Post("/", handlers.ToyImage.Create)
				r.Delete("/{imageID}", handlers.ToyImage.Delete)
			})
		})
	})

	return r
}
