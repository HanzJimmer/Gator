-- name: MarkFeedFetched :exec
UPDATE feeds SET last_fetched_at = $1 WHERE id = $2;
UPDATE feeds SET updated_at = $1 WHERE id = $2;