-- name: GetFeedFollowsForUser :many
SELECT
    users.name AS user_name,
    feeds.name AS feed_name
FROM 
    feed_follows
    JOIN users ON users.id = feed_follows.user_id
    JOIN feeds ON feeds.id = feed_follows.feed_id
WHERE
    feed_follows.user_id = $1;