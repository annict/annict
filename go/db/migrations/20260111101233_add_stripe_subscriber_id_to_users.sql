-- migrate:up
ALTER TABLE users ADD COLUMN stripe_subscriber_id BIGINT;

CREATE INDEX idx_users_stripe_subscriber_id ON users(stripe_subscriber_id);

ALTER TABLE users ADD CONSTRAINT fk_users_stripe_subscriber
    FOREIGN KEY (stripe_subscriber_id) REFERENCES stripe_subscribers(id);

-- migrate:down
ALTER TABLE users DROP CONSTRAINT fk_users_stripe_subscriber;

DROP INDEX idx_users_stripe_subscriber_id;

ALTER TABLE users DROP COLUMN stripe_subscriber_id;
