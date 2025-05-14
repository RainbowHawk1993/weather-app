package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"weather-app/internal/core"
	"weather-app/internal/platform/weatherprovider"
)

type WeatherProvider interface {
	FetchWeather(city string) (*core.Weather, error)
}

type WeatherHandler struct {
	provider WeatherProvider
}

func NewWeatherHandler(p WeatherProvider) *WeatherHandler {
	return &WeatherHandler{provider: p}
}

func (h *WeatherHandler) GetWeather(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	if city == "" {
		http.Error(w, `{"error": "city query parameter is required"}`, http.StatusBadRequest)
		return
	}

	weatherData, err := h.provider.FetchWeather(city)
	if err != nil {
		if errors.Is(err, weatherprovider.ErrCityNotFound) {
			http.Error(w, `{"error": "City not found"}`, http.StatusNotFound)
		} else if errors.Is(err, weatherprovider.ErrAPIRequest) {
			http.Error(w, `{"error": "Failed to fetch weather data from provider"}`, http.StatusInternalServerError)
		} else {
			http.Error(w, `{"error": "An unexpected error occurred"}`, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(weatherData); err != nil {
		log.Printf("Error encoding weather data to JSON: %v", err)
	}
}
