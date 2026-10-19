-- name: CreateOrder :one
INSERT INTO orders (toy_id, phone, message)
VALUES
  ($1, $2, $3)
RETURNING
  id,
  toy_id,
  phone,
  message,
  status,
  created_at,
  updated_at;

-- name: GetOrders :many
SELECT
  id,
  toy_id,
  phone,
  message,
  status,
  created_at,
  updated_at
FROM
  orders
WHERE
  sqlc.narg(status)::text IS NULL
  OR status = sqlc.narg(status)
ORDER BY
  created_at DESC,
  id
LIMIT $1 OFFSET $2;

-- name: GetOrdersByToyID :many
SELECT
  id,
  toy_id,
  phone,
  message,
  status,
  created_at,
  updated_at
FROM
  orders
WHERE
  toy_id = sqlc.arg(toy_id)::uuid
  AND (
    sqlc.narg(status)::text IS NULL
    OR status = sqlc.narg(status)
  )
ORDER BY
  created_at DESC,
  id
LIMIT $1 OFFSET $2;

-- name: GetCustomOrders :many
SELECT
  id,
  toy_id,
  phone,
  message,
  status,
  created_at,
  updated_at
FROM
  orders
WHERE
  toy_id IS NULL
  AND (
    sqlc.narg(status)::text IS NULL
    OR status = sqlc.narg(status)
  )
ORDER BY
  created_at DESC,
  id
LIMIT $1 OFFSET $2;

-- name: GetOrder :one
SELECT
  id,
  toy_id,
  phone,
  message,
  status,
  created_at,
  updated_at
FROM
  orders
WHERE
  id = $1;

-- name: UpdateOrderStatus :one
UPDATE
  orders
SET
  status = $2,
  updated_at = NOW()
WHERE
  id = $1
RETURNING
  id,
  toy_id,
  phone,
  message,
  status,
  created_at,
  updated_at;
