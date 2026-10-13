-- name: CreateImageForToy :one
INSERT INTO toy_images (toy_id, url, position)
VALUES
  ($1, $2, $3)
RETURNING
  id,
  toy_id,
  url,
  position,
  created_at,
  updated_at;

-- name: GetImagesByToyIDs :many
SELECT
  id,
  toy_id,
  url,
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

-- name: GetImageByID :one
SELECT
  id,
  toy_id,
  url,
  position,
  created_at,
  updated_at
FROM
  toy_images
WHERE
  toy_id = $1
  AND id = $2;

-- name: UpdateToyImage :one
UPDATE
  toy_images
SET
  url = $2,
  updated_at = NOW()
WHERE
  id = $1
RETURNING
  id,
  toy_id,
  url,
  position,
  created_at,
  updated_at;

-- name: DeleteToyImage :execrows
DELETE FROM
  toy_images
WHERE
  id = $1;
