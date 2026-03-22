-- migrate:up
CREATE TABLE feature_flags (
    id BIGSERIAL PRIMARY KEY,
    device_token VARCHAR,
    user_id BIGINT REFERENCES users(id),
    name VARCHAR NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CHECK (device_token IS NOT NULL OR user_id IS NOT NULL),
    UNIQUE(device_token, name),
    UNIQUE(user_id, name)
);

-- migrate:down
DROP TABLE IF EXISTS feature_flags;
