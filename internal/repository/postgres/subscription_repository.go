package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
		FROM subscriptions
		WHERE ($1::uuid IS NULL OR user_id = $1)
			AND ($2::text = '' OR service_name = $2)
			AND (
				$3::date IS NULL
				OR $4::date IS NULL
				OR (start_date <= $4 AND (end_date IS NULL OR end_date >= $3))
			)
		ORDER BY created_at DESC
		LIMIT $5 OFFSET $6`

	rows, err := r.db.QueryContext(ctx, query, filterArgs(filter)...)
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

func (r *SubscriptionRepository) Total(ctx context.Context, filter model.SubscriptionFilter) (int, error) {
	query := `
		SELECT COALESCE(SUM(price * (
			(EXTRACT(YEAR FROM AGE(overlap_end, overlap_start))::int * 12)
			+ EXTRACT(MONTH FROM AGE(overlap_end, overlap_start))::int
			+ 1
		)), 0)
		FROM (
			SELECT
				price,
				GREATEST(start_date, $3::date) AS overlap_start,
				LEAST(COALESCE(end_date, $4::date), $4::date) AS overlap_end
			FROM subscriptions
			WHERE ($1::uuid IS NULL OR user_id = $1)
				AND ($2::text = '' OR service_name = $2)
				AND start_date <= $4::date
				AND (end_date IS NULL OR end_date >= $3::date)
		) periods`

	args := filterArgs(filter)
	var total int
	if err := r.db.QueryRowContext(ctx, query, args[0], args[1], args[2], args[3]).Scan(&total); err != nil {
		return 0, fmt.Errorf("total subscriptions: %w", err)
	}

	return total, nil
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

func filterArgs(filter model.SubscriptionFilter) []any {
	var userID any
	if filter.UserID != nil {
		userID = *filter.UserID
	}

	var periodStart any
	if filter.PeriodStart != nil {
		periodStart = *filter.PeriodStart
	}

	var periodEnd any
	if filter.PeriodEnd != nil {
		periodEnd = *filter.PeriodEnd
	}

	return []any{
		userID,
		filter.ServiceName,
		periodStart,
		periodEnd,
		filter.Limit,
		filter.Offset,
	}
}
