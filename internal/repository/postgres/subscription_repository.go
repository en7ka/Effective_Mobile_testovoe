package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/en7ka/Effective_Mobile_testovoe.git/internal/model"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("subscription not found")

type SubscriptionRepository struct {
	db *sql.DB
}

func NewSubscriptionRepository(db *sql.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(ctx context.Context, subscription model.Subscription) (model.Subscription, error) {
	query := `
		INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`

	err := r.db.QueryRowContext(
		ctx,
		query,
		subscription.ID,
		subscription.ServiceName,
		subscription.Price,
		subscription.UserID,
		subscription.StartDate,
		subscription.EndDate,
	).Scan(&subscription.CreatedAt, &subscription.UpdatedAt)
	if err != nil {
		return model.Subscription{}, fmt.Errorf("create subscription: %w", err)
	}

	return subscription, nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (model.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE id = $1`

	subscription, err := scanSubscription(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Subscription{}, ErrNotFound
		}

		return model.Subscription{}, fmt.Errorf("get subscription: %w", err)
	}

	return subscription, nil
}

func (r *SubscriptionRepository) List(ctx context.Context, filter model.SubscriptionFilter) ([]model.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions`

	args := make([]any, 0)
	conditions := make([]string, 0)

	if filter.UserID != nil {
		args = append(args, *filter.UserID)
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", len(args)))
	}

	if filter.ServiceName != "" {
		args = append(args, filter.ServiceName)
		conditions = append(conditions, fmt.Sprintf("service_name = $%d", len(args)))
	}

	if filter.PeriodStart != nil && filter.PeriodEnd != nil {
		args = append(args, *filter.PeriodEnd)
		conditions = append(conditions, fmt.Sprintf("start_date <= $%d", len(args)))

		args = append(args, *filter.PeriodStart)
		conditions = append(conditions, fmt.Sprintf("(end_date IS NULL OR end_date >= $%d)", len(args)))
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}
	defer rows.Close()

	subscriptions := make([]model.Subscription, 0)
	for rows.Next() {
		subscription, err := scanSubscription(rows)
		if err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}

		subscriptions = append(subscriptions, subscription)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return subscriptions, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, subscription model.Subscription) (model.Subscription, error) {
	query := `
		UPDATE subscriptions
		SET service_name = $2,
			price = $3,
			user_id = $4,
			start_date = $5,
			end_date = $6,
			updated_at = now()
		WHERE id = $1
		RETURNING created_at, updated_at`

	err := r.db.QueryRowContext(
		ctx,
		query,
		subscription.ID,
		subscription.ServiceName,
		subscription.Price,
		subscription.UserID,
		subscription.StartDate,
		subscription.EndDate,
	).Scan(&subscription.CreatedAt, &subscription.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Subscription{}, ErrNotFound
		}

		return model.Subscription{}, fmt.Errorf("update subscription: %w", err)
	}

	return subscription, nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM subscriptions WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanSubscription(row scanner) (model.Subscription, error) {
	var subscription model.Subscription
	var endDate sql.NullTime

	err := row.Scan(
		&subscription.ID,
		&subscription.ServiceName,
		&subscription.Price,
		&subscription.UserID,
		&subscription.StartDate,
		&endDate,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)
	if err != nil {
		return model.Subscription{}, err
	}

	if endDate.Valid {
		subscription.EndDate = &endDate.Time
	}

	return subscription, nil
}
