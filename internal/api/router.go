package api

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
		r.Get("/unsubscribe/{token}", sh.Unsubscribe)
	})

	// for serving static html file
	workDir, _ := os.Getwd()
	webDirPath := filepath.Join(workDir, "web")

	if _, err := os.Stat(webDirPath); os.IsNotExist(err) {
		altWebDirPath := filepath.Join(workDir, "..", "web")
		if _, errAlt := os.Stat(altWebDirPath); !os.IsNotExist(errAlt) {
			webDirPath = altWebDirPath
		} else {
			log.Printf("Static files 'web' directory not found at %s or %s. HTML page will not be served.",
				filepath.Join(workDir, "web"), altWebDirPath)
			return r
		}
	}
	log.Printf("Serving static files from: %s", webDirPath)

	fileServer := http.FileServer(http.Dir(webDirPath))

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		fileServer.ServeHTTP(w, r)
	})

	return r
}
