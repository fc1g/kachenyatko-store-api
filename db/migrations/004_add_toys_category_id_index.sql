-- +goose Up
CREATE INDEX toys_category_id_idx ON toys (category_id);

-- +goose Down
DROP INDEX IF EXISTS toys_category_id_idx;
