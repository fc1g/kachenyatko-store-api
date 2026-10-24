package router

import (
	"net/http"
	"strings"
	"time"

	_ "github.com/fc1g/kachenyatko-store-api/docs"
	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/auth"
	"github.com/fc1g/kachenyatko-store-api/internal/category"
	"github.com/fc1g/kachenyatko-store-api/internal/health"
	"github.com/fc1g/kachenyatko-store-api/internal/httpx"
	"github.com/fc1g/kachenyatko-store-api/internal/order"
	"github.com/fc1g/kachenyatko-store-api/internal/session"
	"github.com/fc1g/kachenyatko-store-api/internal/toy"
	"github.com/fc1g/kachenyatko-store-api/internal/user"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

const (
	maxBodyBytes = 1 << 20

	loginRequests = 5
	loginWindow   = time.Minute

	orderRequests = 10
	orderWindow   = 10 * time.Minute

	corsMaxAge = 10 * time.Minute
)

type Handlers struct {
	Health   *health.Handler
	Session  *session.Handler
	User     *user.Handler
	Category *category.Handler
	Toy      *toy.Handler
	ToyImage *toy.ImageHandler
	Order    *order.Handler
}

func New(logger zerolog.Logger, h Handlers, tokens *auth.TokenManager, clientIPHeader, siteURL string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(clientIP(clientIPHeader))
	r.Use(httpx.RequestLogger(logger))
	r.Use(httpx.Recoverer)
	r.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{strings.TrimSuffix(siteURL, "/")},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
		MaxAge:           int(corsMaxAge.Seconds()),
	}).Handler)

	requireAuth := auth.RequireAuth(tokens)
	limitBody := middleware.RequestSize(maxBodyBytes)

	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	r.Get("/livez", h.Health.Live)
	r.Get("/readyz", h.Health.Ready)

	loginLimit := limitByIP(loginRequests, loginWindow, "too many login attempts, try again later")
	orderLimit := limitByIP(orderRequests, orderWindow, "too many orders, try again later")

	r.With(loginLimit, limitBody).Post("/login", h.Session.Login)
	r.Post("/logout", h.Session.Logout)
	r.With(requireAuth).Get("/me", h.Session.Me)

	r.Route("/users", func(r chi.Router) {
		r.Use(requireAuth)

		r.Get("/", h.User.List)
		r.Get("/{userID}", h.User.Get)
		r.With(limitBody).Patch("/{userID}", h.User.Update)
		r.Delete("/{userID}", h.User.Delete)
	})

	r.Route("/categories", func(r chi.Router) {
		r.Get("/", h.Category.List)
		r.Get("/{slug}", h.Category.Get)

		r.Group(func(r chi.Router) {
			r.Use(requireAuth, limitBody)

			r.Post("/", h.Category.Create)
			r.Put("/{categoryID}", h.Category.Update)
			r.Delete("/{categoryID}", h.Category.Delete)
		})
	})

	r.Route("/toys", func(r chi.Router) {
		r.Get("/", h.Toy.List)
		r.Get("/bestsellers", h.Toy.ListBestsellers)
		r.Get("/{toyID}", h.Toy.Get)

		r.Group(func(r chi.Router) {
			r.Use(requireAuth, limitBody)

			r.Post("/", h.Toy.Create)
			r.Put("/{toyID}", h.Toy.Update)
			r.Delete("/{toyID}", h.Toy.Delete)
		})

		r.Route("/{toyID}/images", func(r chi.Router) {
			r.Get("/", h.ToyImage.List)
			r.Get("/{imageID}", h.ToyImage.Get)

			r.Group(func(r chi.Router) {
				r.Use(requireAuth)

				r.With(middleware.RequestSize(toy.MaxImageUploadBytes)).Post("/", h.ToyImage.Create)
				r.Delete("/{imageID}", h.ToyImage.Delete)
			})
		})

		r.Route("/{toyID}/orders", func(r chi.Router) {
			r.With(orderLimit, limitBody).Post("/", h.Order.CreateForToy)
			r.With(requireAuth).Get("/", h.Order.ListByToy)
		})
	})

	r.Route("/orders", func(r chi.Router) {
		r.With(orderLimit, limitBody).Post("/", h.Order.CreateCustom)

		r.Group(func(r chi.Router) {
			r.Use(requireAuth)

			r.Get("/", h.Order.List)
			r.Get("/custom", h.Order.ListCustom)
			r.Get("/{orderID}", h.Order.Get)
			r.With(limitBody).Patch("/{orderID}/status", h.Order.UpdateStatus)
		})
	})

	return r
}

func limitByIP(requests int, window time.Duration, msg string) func(http.Handler) http.Handler {
	return httprate.LimitBy(requests, window,
		func(r *http.Request) (string, error) {
			return httprate.CanonicalizeIP(middleware.GetClientIP(r.Context())), nil
		},
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			httpx.WriteError(w, r, apperr.TooManyRequests(msg, nil))
		}),
	)
}

func clientIP(header string) func(http.Handler) http.Handler {
	if header == "" {
		return middleware.ClientIPFromRemoteAddr
	}
	return middleware.ClientIPFromHeader(header)
}
