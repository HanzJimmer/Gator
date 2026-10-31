-- name: GetFeedsWithUsers :many
SELECT
    users.name AS user_name,
    feeds.name,
    feeds.url
FROM
    feeds
    LEFT JOIN users ON users.id = feeds.user_id;