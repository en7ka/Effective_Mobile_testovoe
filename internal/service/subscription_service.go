package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/en7ka/Effective_Mobile_testovoe.git/internal/model"
	"github.com/en7ka/Effective_Mobile_testovoe.git/internal/repository/postgres"
	"github.com/google/uuid"
)

var (
	ErrNotFound   = errors.New("subscription not found")
	ErrValidation = errors.New("validation error")
)

type Repository interface {
	Create(ctx context.Context, subscription model.Subscription) (model.Subscription, error)
	GetByID(ctx context.Context, id uuid.UUID) (model.Subscription, error)
	List(ctx context.Context, filter model.SubscriptionFilter) ([]model.Subscription, error)
	Update(ctx context.Context, subscription model.Subscription) (model.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type SubscriptionService struct {
	repo Repository
}

func NewSubscriptionService(repo Repository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

func (s *SubscriptionService) Create(ctx context.Context, subscription model.Subscription) (model.Subscription, error) {
	subscription.ID = uuid.New()

	if err := validateSubscription(subscription); err != nil {
		return model.Subscription{}, err
	}

	return s.repo.Create(ctx, subscription)
}

func (s *SubscriptionService) GetByID(ctx context.Context, id uuid.UUID) (model.Subscription, error) {
	subscription, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return model.Subscription{}, mapRepoError(err)
	}

	return subscription, nil
}

func (s *SubscriptionService) List(ctx context.Context, filter model.SubscriptionFilter) ([]model.Subscription, error) {
	return s.repo.List(ctx, filter)
}

func (s *SubscriptionService) Update(ctx context.Context, subscription model.Subscription) (model.Subscription, error) {
	if subscription.ID == uuid.Nil {
		return model.Subscription{}, fmt.Errorf("%w: id is required", ErrValidation)
	}

	if err := validateSubscription(subscription); err != nil {
		return model.Subscription{}, err
	}

	subscription, err := s.repo.Update(ctx, subscription)
	if err != nil {
		return model.Subscription{}, mapRepoError(err)
	}

	return subscription, nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return mapRepoError(err)
	}

	return nil
}

func (s *SubscriptionService) Total(ctx context.Context, filter model.SubscriptionFilter) (int, error) {
	if filter.PeriodStart == nil || filter.PeriodEnd == nil {
		return 0, fmt.Errorf("%w: start_date and end_date are required", ErrValidation)
	}

	if filter.PeriodEnd.Before(*filter.PeriodStart) {
		return 0, fmt.Errorf("%w: end_date must not be before start_date", ErrValidation)
	}

	subscriptions, err := s.repo.List(ctx, filter)
	if err != nil {
		return 0, err
	}

	total := 0
	for _, subscription := range subscriptions {
		from := maxMonth(subscription.StartDate, *filter.PeriodStart)
		to := *filter.PeriodEnd
		if subscription.EndDate != nil {
			to = minMonth(*subscription.EndDate, to)
		}

		if to.Before(from) {
			continue
		}

		total += subscription.Price * monthsCount(from, to)
	}

	return total, nil
}

func validateSubscription(subscription model.Subscription) error {
	if strings.TrimSpace(subscription.ServiceName) == "" {
		return fmt.Errorf("%w: service_name is required", ErrValidation)
	}

	if subscription.Price <= 0 {
		return fmt.Errorf("%w: price must be greater than 0", ErrValidation)
	}

	if subscription.UserID == uuid.Nil {
		return fmt.Errorf("%w: user_id is required", ErrValidation)
	}

	if subscription.StartDate.IsZero() {
		return fmt.Errorf("%w: start_date is required", ErrValidation)
	}

	if subscription.EndDate != nil && subscription.EndDate.Before(subscription.StartDate) {
		return fmt.Errorf("%w: end_date must not be before start_date", ErrValidation)
	}

	return nil
}

func mapRepoError(err error) error {
	if errors.Is(err, postgres.ErrNotFound) {
		return ErrNotFound
	}

	return err
}

func maxMonth(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}

	return b
}

func minMonth(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}

	return b
}

func monthsCount(from, to time.Time) int {
	years := to.Year() - from.Year()
	months := int(to.Month() - from.Month())

	return years*12 + months + 1
}
