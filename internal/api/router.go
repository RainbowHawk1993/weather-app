package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(wh *WeatherHandler, sh *SubscriptionHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		r.Get("/weather", wh.GetWeather)

		r.Post("/subscribe", sh.Subscribe)
		r.Get("/confirm/{token}", sh.ConfirmSubscription)
		// r.Get("/unsubscribe/{token}", sh.Unsubscribe)
	})

	return r
}
