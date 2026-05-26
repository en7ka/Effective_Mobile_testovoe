package model

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID  `validate:"omitempty"`
	ServiceName string     `validate:"notblank"`
	Price       int        `validate:"gt=0"`
	UserID      uuid.UUID  `validate:"notniluuid"`
	StartDate   time.Time  `validate:"required"`
	EndDate     *time.Time `validate:"omitempty,gtefield=StartDate"`
	CreatedAt   time.Time  `validate:"-"`
	UpdatedAt   time.Time  `validate:"-"`
}

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

type SubscriptionFilter struct {
	UserID      *uuid.UUID
	ServiceName string
	PeriodStart *time.Time
	PeriodEnd   *time.Time
	Limit       int
	Offset      int
}
