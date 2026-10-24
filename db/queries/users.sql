-- name: CreateUser :one
INSERT INTO users (email, hashed_password)
VALUES
  ($1, $2)
RETURNING
  id,
  email,
  created_at,
  updated_at;

-- name: ListUsers :many
SELECT
  id,
  email,
  created_at,
  updated_at
FROM
  users
ORDER BY
  created_at,
  id
LIMIT $1 OFFSET $2;

-- name: GetUserByEmail :one
SELECT
  id,
  email,
  hashed_password,
  created_at,
  updated_at
FROM
  users
WHERE
  LOWER(email) = LOWER($1);

-- name: GetUserByID :one
SELECT
  id,
  email,
  created_at,
  updated_at
FROM
  users
WHERE
  id = $1;

-- name: UpdateUserEmail :one
UPDATE
  users
SET
  email = $2,
  updated_at = NOW()
WHERE
  id = $1
RETURNING
  id,
  email,
  created_at,
  updated_at;

-- name: UpdateUserPassword :one
UPDATE
  users
SET
  hashed_password = $2,
  updated_at = NOW()
WHERE
  id = $1
RETURNING
  id,
  email,
  created_at,
  updated_at;

-- name: DeleteUser :execrows
DELETE FROM
  users
WHERE
  id = $1;
