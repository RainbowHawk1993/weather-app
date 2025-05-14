package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"weather-app/internal/api"
	"weather-app/internal/platform/weatherprovider"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	weatherAPIKey := os.Getenv("WEATHERAPI_COM_KEY")
	if weatherAPIKey == "" {
		log.Fatal("Error: WEATHERAPI_COM_KEY environment variable not set.")
	}

	weatherClient := weatherprovider.NewClient(weatherAPIKey)

	weatherHandler := api.NewWeatherHandler(weatherClient)

	router := api.NewRouter(weatherHandler)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Starting server on port %s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v\n", port, err)
	}
}
