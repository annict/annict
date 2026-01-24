-- name: GetGumroadSubscriberByID :one
SELECT * FROM gumroad_subscribers
WHERE id = $1
LIMIT 1;
