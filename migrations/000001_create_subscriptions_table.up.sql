CREATE TABLE subscriptions (
    id UUID PRIMARY KEY,
    service_name TEXT NOT NULL CHECK (btrim(service_name) <> ''),
    price INTEGER NOT NULL CHECK (price > 0),
    user_id UUID NOT NULL,
    start_date DATE NOT NULL CHECK (EXTRACT(DAY FROM start_date) = 1),
    end_date DATE CHECK (end_date IS NULL OR EXTRACT(DAY FROM end_date) = 1),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (end_date IS NULL OR end_date >= start_date)
);

CREATE INDEX idx_subscriptions_user_service
    ON subscriptions (user_id, service_name);

CREATE INDEX idx_subscriptions_period
    ON subscriptions (start_date, end_date);
