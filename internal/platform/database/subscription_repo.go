package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
	"weather-app/internal/core"

	"github.com/jmoiron/sqlx"
)

type SubscriptionRepository interface {
	Create(sub *core.Subscription) error
	FindByEmailAndCity(email, city string) (*core.Subscription, error)
	FindByConfirmationToken(token string) (*core.Subscription, error)
	Confirm(id string) error
	FindByUnsubscribeToken(token string) (*core.Subscription, error)
	Delete(id string) error
}

type PGSubscriptionRepository struct {
	db *sqlx.DB
}

func NewPGSubscriptionRepository(db *sqlx.DB) *PGSubscriptionRepository {
	return &PGSubscriptionRepository{db: db}
}

func (r *PGSubscriptionRepository) Create(sub *core.Subscription) error {
	query := `INSERT INTO subscriptions (id, email, city, frequency, confirmation_token, unsubscribe_token, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	sub.CreatedAt = time.Now().UTC()
	sub.UpdatedAt = time.Now().UTC()

	_, err := r.db.Exec(query, sub.ID, sub.Email, sub.City, sub.Frequency, sub.ConfirmationToken, sub.UnsubscribeToken, sub.CreatedAt, sub.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}
	return nil
}

func (r *PGSubscriptionRepository) FindByEmailAndCity(email, city string) (*core.Subscription, error) {
	var sub core.Subscription
	query := `SELECT id, email, city, frequency, confirmation_token, is_confirmed, unsubscribe_token, created_at, updated_at
              FROM subscriptions WHERE email = $1 AND city = $2`
	err := r.db.Get(&sub, query, email, city)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find subscription by email and city: %w", err)
	}
	return &sub, nil
}

func (r *PGSubscriptionRepository) FindByConfirmationToken(token string) (*core.Subscription, error) {
	var sub core.Subscription
	query := `SELECT id, email, city, frequency, confirmation_token, is_confirmed, unsubscribe_token, created_at, updated_at
              FROM subscriptions WHERE confirmation_token = $1`
	err := r.db.Get(&sub, query, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find subscription by confirmation token: %w", err)
	}
	return &sub, nil
}

func (r *PGSubscriptionRepository) Confirm(id string) error {
	query := `UPDATE subscriptions SET is_confirmed = TRUE, confirmation_token = NULL, updated_at = $1
              WHERE id = $2 AND is_confirmed = FALSE`

	res, err := r.db.Exec(query, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to confirm subscription: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("subscription not found or already confirmed")
	}
	return nil
}

func (r *PGSubscriptionRepository) FindByUnsubscribeToken(token string) (*core.Subscription, error) {
	var sub core.Subscription
	query := `SELECT id, email, city, frequency, confirmation_token, is_confirmed, unsubscribe_token, created_at, updated_at
              FROM subscriptions WHERE unsubscribe_token = $1`
	err := r.db.Get(&sub, query, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find subscription by unsubscribe token: %w", err)
	}
	return &sub, nil
}

func (r *PGSubscriptionRepository) Delete(id string) error {
	query := `DELETE FROM subscriptions WHERE id = $1`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected on delete: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("subscription not found for deletion")
	}
	return nil
}
