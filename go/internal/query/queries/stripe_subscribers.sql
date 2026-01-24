-- name: CreateStripeSubscriber :one
INSERT INTO stripe_subscribers (
    stripe_customer_id,
    stripe_subscription_id,
    stripe_price_id,
    stripe_status,
    stripe_current_period_start,
    stripe_current_period_end,
    stripe_cancel_at,
    stripe_canceled_at,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()
) RETURNING *;

-- name: GetStripeSubscriberByID :one
SELECT * FROM stripe_subscribers
WHERE id = $1
LIMIT 1;

-- name: GetStripeSubscriberByStripeCustomerID :one
SELECT * FROM stripe_subscribers
WHERE stripe_customer_id = $1
LIMIT 1;

-- name: GetStripeSubscriberByStripeSubscriptionID :one
SELECT * FROM stripe_subscribers
WHERE stripe_subscription_id = $1
LIMIT 1;

-- name: UpdateStripeSubscriber :exec
UPDATE stripe_subscribers
SET
    stripe_price_id = $2,
    stripe_status = $3,
    stripe_current_period_start = $4,
    stripe_current_period_end = $5,
    stripe_cancel_at = $6,
    stripe_canceled_at = $7,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateStripeSubscriberStatus :exec
UPDATE stripe_subscribers
SET
    stripe_status = $2,
    updated_at = NOW()
WHERE id = $1;
