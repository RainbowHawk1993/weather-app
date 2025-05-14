package core

import "time"

type Weather struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Description string  `json:"description"`
}

type Subscription struct {
	ID                string    `db:"id" json:"id"` //UUID
	Email             string    `db:"email" json:"email"`
	City              string    `db:"city" json:"city"`
	Frequency         string    `db:"frequency" json:"frequency"`
	ConfirmationToken *string   `db:"confirmation_token" json:"-"`
	IsConfirmed       bool      `db:"is_confirmed" json:"confirmed"`
	UnsubscribeToken  string    `db:"unsubscribe_token" json:"-"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
}

type SubscriptionRequest struct {
	Email     string `form:"email" json:"email"`
	City      string `form:"city" json:"city"`
	Frequency string `form:"frequency" json:"frequency"`
}
