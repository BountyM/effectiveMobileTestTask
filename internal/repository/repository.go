package repository

import (
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	Subscription
}

func New(db *sqlx.DB) *Repository {
	return &Repository{
		Subscription: NewSubscriptionPostgres(db),
	}
}
