-- name: CreateCategory :one
INSERT INTO categories (name, slug)
VALUES
  ($1, $2)
RETURNING
  *;

-- name: GetCategories :many
SELECT
  id,
  name,
  slug,
  created_at,
  updated_at
FROM
  categories
ORDER BY
  name;
