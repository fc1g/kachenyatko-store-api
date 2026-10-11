-- +goose Up
CREATE TABLE toy_images (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  toy_id UUID NOT NULL REFERENCES toys (id) ON DELETE CASCADE,
  url TEXT NOT NULL,
  position INT NOT NULL,
  UNIQUE(toy_id, position),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE toy_images;
