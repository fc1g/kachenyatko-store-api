-- name: CreateToy :one
INSERT INTO toys (
  category_id, name, description, size,
  materials
)
VALUES
  ($1, $2, $3, $4, $5)
RETURNING
  id,
  category_id,
  name,
  description,
  size,
  materials,
  created_at,
  updated_at;

-- name: GetToys :many
SELECT
  id,
  category_id,
  name,
  description,
  size,
  materials,
  created_at,
  updated_at
FROM
  toys
ORDER BY
  created_at DESC;
