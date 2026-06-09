-- name: AddFeed :exec
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
);

-- name: ListFeeds :many
SELECT feeds.name, feeds.url, users.name as creator
FROM feeds
JOIN users
    ON feeds.user_id = users.id;

-- name: FeedByURL :one
SELECT * FROM feeds WHERE url = $1;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET updated_at = NOW(), last_fetched_at = NOW()
WHERE id = $1;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds

ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;