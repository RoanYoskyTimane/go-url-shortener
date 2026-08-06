-- name: CreateURL :one
INSERT INTO urls (short_code, original_url)
VALUES ($1, $2)
    RETURNING id, short_code, original_url, access_count, created_at, updated_at;

-- name: GetURLByShortCode :one
SELECT id, short_code, original_url, access_count, created_at, updated_at
FROM urls
WHERE short_code = $1 LIMIT 1;

-- name: IncrementAccessCount :exec
UPDATE urls
SET access_count = access_count + 1,
    updated_at = NOW()
WHERE short_code = $1;

-- name: DeleteURLByShortCode :exec
DELETE FROM urls
WHERE short_code = $1;