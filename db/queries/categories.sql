-- name: CreateCategory :one
INSERT INTO categories (name, slug)
VALUES
  ($1, $2)
RETURNING
  id,
  name,
  slug,
  created_at,
  updated_at;

-- name: UpsertCategory :one
INSERT INTO categories (name, slug)
VALUES
  ($1, $2) ON CONFLICT (slug) DO
UPDATE
SET
  slug = EXCLUDED.slug
RETURNING
  id,
  name,
  slug,
  created_at,
  updated_at;

-- name: ListCategories :many
SELECT
  id,
  name,
  slug,
  created_at,
  updated_at
FROM
  categories
ORDER BY
  name
LIMIT $1 OFFSET $2;

-- name: CountCategories :one
SELECT
  COUNT(*)
FROM
  categories;

-- name: GetCategory :one
SELECT
  id,
  name,
  slug,
  created_at,
  updated_at
FROM
  categories
WHERE
  id = $1;

-- name: GetCategoryBySlug :one
SELECT
  id,
  name,
  slug,
  created_at,
  updated_at
FROM
  categories
WHERE
  slug = $1;

-- name: UpdateCategory :one
UPDATE
  categories
SET
  name = $2,
  slug = $3,
  updated_at = NOW()
WHERE
  id = $1
RETURNING
  id,
  name,
  slug,
  created_at,
  updated_at;

-- name: DeleteCategory :execrows
DELETE FROM
  categories
WHERE
  id = $1;
