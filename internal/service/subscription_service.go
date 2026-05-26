package service

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/en7ka/Effective_Mobile_testovoe.git/internal/model"
	"github.com/en7ka/Effective_Mobile_testovoe.git/internal/repository/postgres"
	"github.com/go-playground/validator/v10"
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
	Total(ctx context.Context, filter model.SubscriptionFilter) (int, error)
	Update(ctx context.Context, subscription model.Subscription) (model.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type SubscriptionService struct {
	repo     Repository
	validate *validator.Validate
}

func NewSubscriptionService(repo Repository) *SubscriptionService {
	validate := validator.New(validator.WithRequiredStructEnabled())
	_ = validate.RegisterValidation("notblank", validateNotBlank)
	_ = validate.RegisterValidation("notniluuid", validateNotNilUUID)

	return &SubscriptionService{
		repo:     repo,
		validate: validate,
	}
}

func (s *SubscriptionService) Create(ctx context.Context, subscription model.Subscription) (model.Subscription, error) {
	subscription.ID = uuid.New()

	if err := s.validateSubscription(subscription); err != nil {
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
	normalizePagination(&filter)

	return s.repo.List(ctx, filter)
}

func (s *SubscriptionService) Update(ctx context.Context, subscription model.Subscription) (model.Subscription, error) {
	if subscription.ID == uuid.Nil {
		return model.Subscription{}, fmt.Errorf("%w: id is required", ErrValidation)
	}

	if err := s.validateSubscription(subscription); err != nil {
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

	return s.repo.Total(ctx, filter)
}

func (s *SubscriptionService) validateSubscription(subscription model.Subscription) error {
	if err := s.validate.Struct(subscription); err != nil {
		return fmt.Errorf("%w: %s", ErrValidation, validationMessage(err))
	}
	return nil
}

func normalizePagination(filter *model.SubscriptionFilter) {
	if filter.Limit <= 0 {
		filter.Limit = model.DefaultLimit
	}

	if filter.Limit > model.MaxLimit {
		filter.Limit = model.MaxLimit
	}

	if filter.Offset < 0 {
		filter.Offset = 0
	}
}

func validateNotBlank(fl validator.FieldLevel) bool {
	return strings.TrimSpace(fl.Field().String()) != ""
}

func validateNotNilUUID(fl validator.FieldLevel) bool {
	value := fl.Field()
	if value.Kind() == reflect.Array {
		id, ok := value.Interface().(uuid.UUID)
		return ok && id != uuid.Nil
	}

	return false
}

func validationMessage(err error) string {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) || len(validationErrors) == 0 {
		return err.Error()
	}

	fieldError := validationErrors[0]
	switch fieldError.Field() {
	case "ServiceName":
		return "service_name is required"
	case "Price":
		return "price must be greater than 0"
	case "UserID":
		return "user_id is required"
	case "StartDate":
		return "start_date is required"
	case "EndDate":
		return "end_date must not be before start_date"
	default:
		return fieldError.Field() + " is invalid"
	}
}

func mapRepoError(err error) error {
	if errors.Is(err, postgres.ErrNotFound) {
		return ErrNotFound
	}

	return err
}
