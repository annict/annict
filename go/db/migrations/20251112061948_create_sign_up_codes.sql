-- migrate:up
CREATE TABLE sign_up_codes (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    code_digest VARCHAR(255) NOT NULL,  -- bcryptでハッシュ化された確認コード
    attempts INTEGER NOT NULL DEFAULT 0,  -- 検証試行回数
    used_at TIMESTAMP WITH TIME ZONE,  -- 使用日時（NULL = 未使用）
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,  -- 有効期限（15分）
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sign_up_codes_email ON sign_up_codes(email);
CREATE INDEX idx_sign_up_codes_expires_at ON sign_up_codes(expires_at);

-- migrate:down
DROP TABLE IF EXISTS sign_up_codes;

