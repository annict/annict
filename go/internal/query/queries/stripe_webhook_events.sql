-- name: CreateStripeWebhookEvent :one
INSERT INTO stripe_webhook_events (
    stripe_event_id,
    event_type,
    payload,
    status,
    received_at
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetStripeWebhookEventByStripeEventID :one
SELECT * FROM stripe_webhook_events
WHERE stripe_event_id = $1
LIMIT 1;

-- name: UpdateStripeWebhookEventStatus :exec
UPDATE stripe_webhook_events
SET status = $2,
    error_message = $3,
    processed_at = $4,
    updated_at = NOW()
WHERE id = $1;
