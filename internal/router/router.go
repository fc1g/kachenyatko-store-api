package router

import (
	"net/http"
	"strings"

	_ "github.com/fc1g/kachenyatko-store-api/docs"
	"github.com/fc1g/kachenyatko-store-api/internal/handler"
	"github.com/fc1g/kachenyatko-store-api/internal/httpmw"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	httpSwagger "github.com/swaggo/http-swagger/v2"
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

type toyImageHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	GetAll(w http.ResponseWriter, r *http.Request)
	GetOne(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

type orderHandler interface {
	CreateForToy(w http.ResponseWriter, r *http.Request)
	CreateCustom(w http.ResponseWriter, r *http.Request)
	GetAll(w http.ResponseWriter, r *http.Request)
	GetAllCustomOrders(w http.ResponseWriter, r *http.Request)
	GetAllOrdersByToyID(w http.ResponseWriter, r *http.Request)
	GetOne(w http.ResponseWriter, r *http.Request)
	UpdateStatus(w http.ResponseWriter, r *http.Request)
}

type Handlers struct {
	Health   healthHandler
	Category categoryHandler
	Toy      toyHandler
	ToyImage toyImageHandler
	Order    orderHandler
}

func New(logger zerolog.Logger, handlers Handlers, siteURL string) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(httpmw.RequestLogger(logger))
	r.Use(middleware.Recoverer)
	r.Use(cors.New(cors.Options{
		AllowedOrigins: []string{strings.TrimSuffix(siteURL, "/")},
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	}).Handler)

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
				r.With(middleware.RequestSize(handler.MaxImageUploadBytes)).Post("/", handlers.ToyImage.Create)
				r.Delete("/{imageID}", handlers.ToyImage.Delete)
			})
		})

		r.Route("/{toyID}/orders", func(r chi.Router) {
			r.With(middleware.RequestSize(maxBodyBytes)).Post("/", handlers.Order.CreateForToy)

			r.Group(func(r chi.Router) {
				// TODO: r.use(requireAdmin)
				r.Get("/", handlers.Order.GetAllOrdersByToyID)
			})
		})
	})

	r.Route("/orders", func(r chi.Router) {
		r.With(middleware.RequestSize(maxBodyBytes)).Post("/", handlers.Order.CreateCustom)

		r.Group(func(r chi.Router) {
			// TODO: r.use(requireAdmin)
			r.Get("/", handlers.Order.GetAll)
			r.Get("/custom", handlers.Order.GetAllCustomOrders)
			r.Get("/{orderID}", handlers.Order.GetOne)
			r.With(middleware.RequestSize(maxBodyBytes)).Patch("/{orderID}/status", handlers.Order.UpdateStatus)
		})
	})

	return r
}
