-- name: CreateImageForToy :one
INSERT INTO toy_images (toy_id, url, position)
VALUES
  ($1, $2, $3)
RETURNING
  *;

-- name: GetImagesByToyID :many
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
ORDER BY
  position;
