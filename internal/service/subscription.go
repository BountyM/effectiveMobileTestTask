package service

import (
	"fmt"

	"github.com/BountyM/effectiveMobileTestTask/internal/models"
	"github.com/BountyM/effectiveMobileTestTask/internal/repository"
	"github.com/google/uuid"
)

type SubscriptionService struct {
	repository repository.Repository
}

func newSubscriptionService(repository repository.Repository) *SubscriptionService {
	return &SubscriptionService{repository: repository}
}

type Subscription interface {
	Create(subscription models.Subscription) (uuid.UUID, error)
	Get(params models.SubscriptionParams) ([]models.Subscription, error)
	Delete(uuid.UUID) error
	Update(uuid uuid.UUID, subscription models.Subscription) error
	GetCost(params models.SubscriptionParams) (int64, error)
}

func (s *SubscriptionService) Create(subscription models.Subscription) (uuid.UUID, error) {
	res, err := s.repository.Create(subscription)
	if err != nil {

		return uuid.Nil, fmt.Errorf("SubscriptionService Create() %w", err)
	}
	return res, err
}

func (s *SubscriptionService) Get(params models.SubscriptionParams) ([]models.Subscription, error) {
	res, err := s.repository.Get(params)
	if err != nil {

		return nil, fmt.Errorf("SubscriptionService Get() %w", err)
	}
	return res, err
}

func (s *SubscriptionService) Delete(id uuid.UUID) error {
	err := s.repository.Delete(id)
	if err != nil {

		return fmt.Errorf("SubscriptionService Delete() %w", err)
	}
	return err
}

func (s *SubscriptionService) Update(uuid uuid.UUID, subscription models.Subscription) error {
	err := s.repository.Update(uuid, subscription)
	if err != nil {

		return fmt.Errorf("SubscriptionService Update() %w", err)
	}
	return err
}

func (s *SubscriptionService) GetCost(params models.SubscriptionParams) (int64, error) {
	res, err := s.repository.GetCost(params)
	if err != nil {
		return 0, fmt.Errorf("SubscriptionService GetCost() %w", err)
	}
	return res, err
}
