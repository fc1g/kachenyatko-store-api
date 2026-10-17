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
  sqlc.embed(toys),
  sqlc.embed(categories)
FROM
  toys
  JOIN categories ON categories.id = toys.category_id
ORDER BY
  toys.created_at DESC,
  toys.id
LIMIT $1 OFFSET $2;

-- name: GetToy :one
SELECT
  sqlc.embed(toys),
  sqlc.embed(categories)
FROM
  toys
  JOIN categories ON categories.id = toys.category_id
WHERE
  toys.id = $1;

-- name: ToyExists :one
SELECT
  EXISTS (
    SELECT
      1
    FROM
      toys
    WHERE
      id = $1
  );

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
