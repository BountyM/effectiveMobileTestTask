package models

import (
	"time"

	"github.com/google/uuid"
)

const SubscriptionTable = "subscription"

// Subscription model
// @name Subscription
type Subscription struct {
	ID          uuid.UUID  `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int64      `json:"price"`
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}

type SubscriptionParams struct {
	Page        int
	Limit       int
	UserID      *uuid.UUID
	ServiceName string
	StartDate   time.Time
	EndDate     time.Time
}
