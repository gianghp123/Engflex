-- +goose Up
CREATE TABLE plans (
    plan        text PRIMARY KEY,
    months      integer NOT NULL,
    amount      bigint NOT NULL,
    currency    text NOT NULL DEFAULT 'VND',
    is_active   boolean NOT NULL DEFAULT true,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- Seed with the prices that previously lived in the Go planCatalog map.
INSERT INTO plans (plan, months, amount, currency) VALUES
 ('1m', 1, 99000, 'VND'),
 ('3m', 3, 249000, 'VND'),
 ('12m', 12, 799000, 'VND');

CREATE TABLE premium_features (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    key         text NOT NULL UNIQUE,
    label       text NOT NULL,
    description text NOT NULL DEFAULT '',
    is_active   boolean NOT NULL DEFAULT true,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS premium_features;
DROP TABLE IF EXISTS plans;
