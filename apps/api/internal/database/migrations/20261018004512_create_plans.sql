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

-- +goose Down
DROP TABLE IF EXISTS plans;
