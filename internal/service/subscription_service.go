// internal/service/subscription_service.go
package service

import (
	"errors"
	"fmt"
	"log"
	"weather-app/internal/core"
	"weather-app/internal/platform/database"
	"weather-app/internal/platform/email"

	"github.com/google/uuid"
)

var (
	ErrSubscriptionAlreadyExists = errors.New("email already subscribed to this city")
	ErrSubscriptionNotFound      = errors.New("subscription not found")
	ErrInvalidToken              = errors.New("invalid or expired token")
	ErrAlreadyConfirmed          = errors.New("subscription already confirmed")
)

type SubscriptionService struct {
	repo       database.SubscriptionRepository
	emailer    email.Service
	appBaseURL string
}

func NewSubscriptionService(
	repo database.SubscriptionRepository,
	emailer email.Service,
	appBaseURL string,
) *SubscriptionService {
	return &SubscriptionService{
		repo:       repo,
		emailer:    emailer,
		appBaseURL: appBaseURL,
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

	log.Printf("Subscription created for %s, city %s. Confirmation token: %s", newSub.Email, newSub.City, confirmationToken)
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
