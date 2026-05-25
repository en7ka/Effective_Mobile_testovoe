package model

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   time.Time
	EndDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SubscriptionFilter struct {
	UserID      *uuid.UUID
	ServiceName string
	PeriodStart *time.Time
	PeriodEnd   *time.Time
}
