-- name: CreateToyImage :one
INSERT INTO toy_images (toy_id, storage_key, position)
SELECT
  sqlc.arg(toy_id):: uuid,
  sqlc.arg(storage_key):: text,
  COALESCE(
    MAX(position),
    -1
  ) + 1
FROM
  toy_images
WHERE
  toy_id = sqlc.arg(toy_id):: uuid
RETURNING
  id,
  toy_id,
  storage_key,
  position,
  created_at,
  updated_at;

-- name: GetToyImagesByToyIDs :many
SELECT
  id,
  toy_id,
  storage_key,
  position,
  created_at,
  updated_at
FROM
  toy_images
WHERE
  toy_id = ANY(sqlc.arg(toy_ids)::uuid[])
ORDER BY
  toy_id,
  position;

-- name: GetToyImage :one
SELECT
  id,
  toy_id,
  storage_key,
  position,
  created_at,
  updated_at
FROM
  toy_images
WHERE
  toy_id = $1
  AND id = $2;

-- name: DeleteToyImage :one
DELETE FROM
  toy_images
WHERE
  toy_id = $1
  AND id = $2
RETURNING
  storage_key;

-- name: DeleteToyImagesByToyID :many
DELETE FROM
  toy_images
WHERE
  toy_id = $1
RETURNING
  storage_key;
