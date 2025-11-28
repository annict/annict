-- migrate:up
CREATE TABLE email_login_codes (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_digest VARCHAR(255) NOT NULL,  -- bcryptでハッシュ化された6桁コード
    attempts INT NOT NULL DEFAULT 0,     -- 試行回数
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,  -- 有効期限
    used_at TIMESTAMP WITH TIME ZONE,    -- 使用済み時刻（NULL = 未使用）
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_login_codes_user_id ON email_login_codes(user_id);
CREATE INDEX idx_email_login_codes_expires_at ON email_login_codes(expires_at);
CREATE INDEX idx_email_login_codes_used_at ON email_login_codes(used_at);

-- migrate:down
DROP TABLE IF EXISTS email_login_codes;
