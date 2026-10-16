-- +goose Up
CREATE TABLE toy_images (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  toy_id UUID NOT NULL REFERENCES toys (id) ON DELETE CASCADE,
  storage_key TEXT NOT NULL UNIQUE,
  position INT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(toy_id, position)
);

-- +goose Down
DROP TABLE IF EXISTS toy_images;
