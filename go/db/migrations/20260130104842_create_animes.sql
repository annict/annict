-- migrate:up
CREATE TABLE animes (
    id BIGSERIAL PRIMARY KEY,
    parent_id BIGINT REFERENCES animes(id),

    -- 共通属性
    title VARCHAR NOT NULL,
    title_kana VARCHAR,
    title_alter VARCHAR,

    -- 集計・評価カラム
    ratings_count INTEGER NOT NULL DEFAULT 0,
    satisfaction_rate NUMERIC(5, 2),
    score NUMERIC(5, 2),

    -- 状態管理
    deleted_at TIMESTAMP WITH TIME ZONE,
    hidden_at TIMESTAMP WITH TIME ZONE,

    -- タイムスタンプ
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX index_animes_on_parent_id ON animes(parent_id);

-- migrate:down
DROP TABLE IF EXISTS animes;
