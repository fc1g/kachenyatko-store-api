-- +goose Up
CREATE TABLE orders (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  toy_id UUID REFERENCES toys (id) ON DELETE RESTRICT,
  phone TEXT NOT NULL,
  message TEXT,
  details TEXT,
  status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'closed')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK (toy_id IS NOT NULL OR (details IS NOT NULL AND details <> ''))
);

CREATE INDEX orders_toy_id_idx ON orders (toy_id);

-- +goose Down
DROP TABLE IF EXISTS orders;
