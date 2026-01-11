-- migrate:up
CREATE TABLE stripe_webhook_events (
    id BIGSERIAL PRIMARY KEY,
    stripe_event_id VARCHAR(255) NOT NULL,
    stripe_event_type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    error_message TEXT,
    received_at TIMESTAMP WITH TIME ZONE NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_stripe_webhook_events_stripe_event_id ON stripe_webhook_events(stripe_event_id);
CREATE INDEX idx_stripe_webhook_events_stripe_event_type ON stripe_webhook_events(stripe_event_type);
CREATE INDEX idx_stripe_webhook_events_status ON stripe_webhook_events(status);
CREATE INDEX idx_stripe_webhook_events_received_at ON stripe_webhook_events(received_at);

-- migrate:down
DROP TABLE IF EXISTS stripe_webhook_events;
