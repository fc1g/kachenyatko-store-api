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
  created_at DESC, id
LIMIT $1 OFFSET $2;

-- name: GetToy :one
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
WHERE
  id = $1;

-- name: UpdateToy :one
UPDATE
  toys
SET
  category_id = $2,
  name = $3,
  description = $4,
  size = $5,
  materials = $6,
  updated_at = NOW()
WHERE
  id = $1
RETURNING
  id,
  category_id,
  name,
  description,
  size,
  materials,
  created_at,
  updated_at;

-- name: DeleteToy :execrows
DELETE FROM
  toys
WHERE
  id = $1;
