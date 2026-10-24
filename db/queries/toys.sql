-- name: LockToy :one
SELECT
  id
FROM
  toys
WHERE
  id = $1 FOR NO KEY UPDATE;

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

-- name: ListToys :many
SELECT
  sqlc.embed(toys),
  sqlc.embed(categories)
FROM
  toys
  JOIN categories ON categories.id = toys.category_id
WHERE
  sqlc.narg(category_slug)::text IS NULL
  OR categories.slug = sqlc.narg(category_slug)
ORDER BY
  toys.created_at DESC,
  toys.id
LIMIT $1 OFFSET $2;

-- name: ListBestsellers :many
SELECT
  sqlc.embed(toys),
  sqlc.embed(categories)
FROM
  toys
  JOIN categories ON categories.id = toys.category_id
  LEFT JOIN (
    SELECT
      toy_id,
      COUNT(*) AS orders_count
    FROM
      orders
    WHERE
      toy_id IS NOT NULL
    GROUP BY
      toy_id
  ) AS order_counts ON order_counts.toy_id = toys.id
WHERE
  sqlc.narg(category_slug)::text IS NULL
  OR categories.slug = sqlc.narg(category_slug)
ORDER BY
  COALESCE(order_counts.orders_count, 0) DESC,
  toys.created_at DESC,
  toys.id
LIMIT $1 OFFSET $2;

-- name: CountToys :one
SELECT
  COUNT(*)
FROM
  toys
  JOIN categories ON categories.id = toys.category_id
WHERE
  sqlc.narg(category_slug)::text IS NULL
  OR categories.slug = sqlc.narg(category_slug);

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
