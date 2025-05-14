package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"weather-app/internal/core"
	"weather-app/internal/platform/weatherprovider"
	"weather-app/internal/service"

	"github.com/go-chi/chi/v5"
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

// GetWeather handles GET /api/weather
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

type SubscriptionHandler struct {
	subService *service.SubscriptionService
}

func NewSubscriptionHandler(ss *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{subService: ss}
}

// Subscribe handles POST /api/subscribe
func (h *SubscriptionHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, `{"error": "Failed to parse form data"}`, http.StatusBadRequest)
		return
	}

	req := core.SubscriptionRequest{
		Email:     r.FormValue("email"),
		City:      r.FormValue("city"),
		Frequency: r.FormValue("frequency"),
	}

	if req.Email == "" || req.City == "" || req.Frequency == "" {
		http.Error(w, `{"error": "email, city, and frequency are required"}`, http.StatusBadRequest)
		return
	}
	if req.Frequency != "hourly" && req.Frequency != "daily" {
		http.Error(w, `{"error": "frequency must be 'hourly' or 'daily'"}`, http.StatusBadRequest)
		return
	}

	err := h.subService.CreateSubscription(req)
	if err != nil {
		log.Printf("Subscribe handler error: %v", err)
		if errors.Is(err, service.ErrSubscriptionAlreadyExists) {
			http.Error(w, `{"error": "Email already subscribed for this city"}`, http.StatusConflict)
		} else if err.Error() == "invalid frequency" {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		} else {
			http.Error(w, `{"error": "Failed to create subscription"}`, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Subscription successful. Confirmation email sent."})
}

func (h *SubscriptionHandler) ConfirmSubscription(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		http.Error(w, `{"error": "Token is required"}`, http.StatusBadRequest)
		return
	}

	err := h.subService.ConfirmSubscription(token)
	if err != nil {
		log.Printf("ConfirmSubscription handler error for token %s: %v", token, err)
		if errors.Is(err, service.ErrSubscriptionNotFound) || errors.Is(err, service.ErrInvalidToken) {
			http.Error(w, `{"error": "Invalid or expired token"}`, http.StatusNotFound)
		} else if errors.Is(err, service.ErrAlreadyConfirmed) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"message": "Subscription already confirmed."})
		} else {
			http.Error(w, `{"error": "Failed to confirm subscription"}`, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Subscription confirmed successfully"})
}
