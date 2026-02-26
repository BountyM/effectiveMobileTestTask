package service

import (
	"github.com/BountyM/effectiveMobileTestTask/internal/repository"
)

type Service struct {
	Subscription
}

func New(repository *repository.Repository) *Service {
	return &Service{
		Subscription: newSubscriptionService(*repository),
	}
}
