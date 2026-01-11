-- migrate:up
CREATE TABLE stripe_subscribers (
    id BIGSERIAL PRIMARY KEY,
    stripe_customer_id VARCHAR(255) NOT NULL,
    stripe_subscription_id VARCHAR(255) NOT NULL,
    stripe_price_id VARCHAR(255) NOT NULL,
    stripe_status VARCHAR(50) NOT NULL,
    stripe_current_period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    stripe_current_period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    stripe_cancel_at TIMESTAMP WITH TIME ZONE,
    stripe_canceled_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_stripe_subscribers_stripe_customer_id ON stripe_subscribers(stripe_customer_id);
CREATE UNIQUE INDEX idx_stripe_subscribers_stripe_subscription_id ON stripe_subscribers(stripe_subscription_id);
CREATE INDEX idx_stripe_subscribers_stripe_status ON stripe_subscribers(stripe_status);

-- migrate:down
DROP TABLE IF EXISTS stripe_subscribers;

