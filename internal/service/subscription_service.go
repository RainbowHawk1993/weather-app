package service

import (
	"errors"
	"fmt"
	"log"
	"time"
	"weather-app/internal/core"
	"weather-app/internal/platform/database"
	"weather-app/internal/platform/email"
	"weather-app/internal/platform/weatherprovider"

	"github.com/google/uuid"
)

var (
	ErrSubscriptionAlreadyExists = errors.New("email already subscribed to this city")
	ErrSubscriptionNotFound      = errors.New("subscription not found")
	ErrInvalidToken              = errors.New("invalid or expired token")
	ErrAlreadyConfirmed          = errors.New("subscription already confirmed")
)

type SubscriptionService struct {
	repo            database.SubscriptionRepository
	emailer         email.Service
	weatherProvider weatherprovider.WeatherProvider
	appBaseURL      string
}

func NewSubscriptionService(
	repo database.SubscriptionRepository,
	emailer email.Service,
	weatherProvider weatherprovider.WeatherProvider,
	appBaseURL string,
) *SubscriptionService {
	return &SubscriptionService{
		repo:            repo,
		emailer:         emailer,
		weatherProvider: weatherProvider,
		appBaseURL:      appBaseURL,
	}
}

func (s *SubscriptionService) CreateSubscription(req core.SubscriptionRequest) error {
	if req.Frequency != "hourly" && req.Frequency != "daily" {
		return fmt.Errorf("invalid frequency: %s. Must be 'hourly' or 'daily'", req.Frequency)
	}

	existingSub, err := s.repo.FindByEmailAndCity(req.Email, req.City)
	if err != nil {
		log.Printf("Error checking for existing subscription: %v", err)
		return fmt.Errorf("could not process subscription request")
	}
	if existingSub != nil {
		if existingSub.IsConfirmed {
			return ErrSubscriptionAlreadyExists
		}
		log.Printf("Subscription exists for %s in %s but not confirmed. ID: %s", req.Email, req.City, existingSub.ID)
		return ErrSubscriptionAlreadyExists
	}

	confirmationToken := uuid.NewString()
	unsubscribeToken := uuid.NewString()

	newSub := &core.Subscription{
		ID:                uuid.NewString(),
		Email:             req.Email,
		City:              req.City,
		Frequency:         req.Frequency,
		ConfirmationToken: &confirmationToken,
		IsConfirmed:       false,
		UnsubscribeToken:  unsubscribeToken,
	}

	if err := s.repo.Create(newSub); err != nil {
		log.Printf("Error creating subscription in DB: %v", err)
		return fmt.Errorf("could not save subscription")
	}

	confirmationLink := fmt.Sprintf("%s/api/confirm/%s", s.appBaseURL, confirmationToken)
	if err := s.emailer.SendConfirmationEmail(newSub.Email, newSub.City, confirmationLink); err != nil {
		log.Printf("Failed to send confirmation email to %s: %v", newSub.Email, err)
	}

	log.Printf("Subscription created for %s, city %s. Confirmation token: %s. Unsubscribe token: %s.",
		newSub.Email, newSub.City, confirmationToken, unsubscribeToken)
	return nil
}

func (s *SubscriptionService) ConfirmSubscription(token string) error {
	if _, err := uuid.Parse(token); err != nil {
		return ErrInvalidToken
	}

	sub, err := s.repo.FindByConfirmationToken(token)
	if err != nil {
		log.Printf("Error finding subscription by confirmation token %s: %v", token, err)
		return fmt.Errorf("database error during confirmation")
	}

	if sub == nil {
		return ErrSubscriptionNotFound
	}

	if sub.IsConfirmed {
		return ErrAlreadyConfirmed
	}

	if err := s.repo.Confirm(sub.ID); err != nil {
		if errors.Is(err, errors.New("subscription not found or already confirmed")) {
			return ErrAlreadyConfirmed
		}
		log.Printf("Error confirming subscription ID %s: %v", sub.ID, err)
		return fmt.Errorf("could not confirm subscription")
	}

	log.Printf("Subscription ID %s confirmed for email %s", sub.ID, sub.Email)
	return nil
}

func (s *SubscriptionService) Unsubscribe(token string) error {
	if _, err := uuid.Parse(token); err != nil {
		return ErrInvalidToken
	}

	sub, err := s.repo.FindByUnsubscribeToken(token)
	if err != nil {
		log.Printf("Error finding subscription by unsubscribe token %s: %v", token, err)
		return fmt.Errorf("database error during unsubscribe lookup")
	}

	if sub == nil {
		return ErrSubscriptionNotFound
	}

	if err := s.repo.Delete(sub.ID); err != nil {
		if err.Error() == "subscription not found for deletion" {
			return ErrSubscriptionNotFound
		}
		log.Printf("Error deleting subscription ID %s: %v", sub.ID, err)
		return fmt.Errorf("could not process unsubscription")
	}

	log.Printf("Subscription ID %s (email: %s, city: %s) unsubscribed successfully.", sub.ID, sub.Email, sub.City)
	return nil
}
func (s *SubscriptionService) SendWeatherUpdates() {
	log.Println("Scheduler: Running SendWeatherUpdates job.")
	now := time.Now().UTC()

	confirmedSubs, err := s.repo.GetAllConfirmed()
	if err != nil {
		log.Printf("Scheduler: Error fetching confirmed subscriptions: %v", err)
		return
	}

	if len(confirmedSubs) == 0 {
		log.Println("Scheduler: No confirmed subscriptions to process.")
		return
	}

	log.Printf("Scheduler: Processing %d confirmed subscriptions at %s.", len(confirmedSubs), now.Format(time.RFC3339))

	for _, sub := range confirmedSubs {
		isDue := false
		switch sub.Frequency {
		case "hourly":
			if now.Minute() == 0 { // Start of the hour
				isDue = true
			}

		case "daily":
			targetHourDaily := 8 // 8 AM UTC
			if now.Hour() == targetHourDaily && now.Minute() == 0 {
				isDue = true
			}
		default:
			log.Printf("Scheduler: Unknown frequency '%s' for subscription ID %s. Skipping.", sub.Frequency, sub.ID)
			continue
		}

		if !isDue {
			continue
		}

		log.Printf("Scheduler: Update DUE for %s (%s) in %s.", sub.Email, sub.Frequency, sub.City)

		weatherData, err := s.weatherProvider.FetchWeather(sub.City)
		if err != nil {
			log.Printf("Scheduler: Failed to fetch weather for %s (subscriber %s): %v", sub.City, sub.Email, err)
			continue
		}

		weatherInfo := fmt.Sprintf(
			"Current weather in %s:\nTemperature: %.1f°C\nHumidity: %.0f%%\nDescription: %s",
			sub.City, weatherData.Temperature, weatherData.Humidity, weatherData.Description,
		)
		unsubscribeLink := fmt.Sprintf("%s/api/unsubscribe/%s", s.appBaseURL, sub.UnsubscribeToken)

		if err := s.emailer.SendWeatherUpdateEmail(sub.Email, sub.City, weatherInfo, unsubscribeLink); err != nil {
			log.Printf("Scheduler: Failed to send weather update to %s for city %s: %v", sub.Email, sub.City, err)
		} else {
			log.Printf("Scheduler: Successfully sent weather update to %s for city %s.", sub.Email, sub.City)
		}
	}
	log.Println("Scheduler: Finished SendWeatherUpdates job run.")
}
