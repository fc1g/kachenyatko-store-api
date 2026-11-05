-- +goose Up
ALTER TABLE orders
ADD COLUMN notified_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE orders
DROP COLUMN IF EXISTS notified_at;
