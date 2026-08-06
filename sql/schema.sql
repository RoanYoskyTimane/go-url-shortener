CREATE TABLE IF NOT EXISTS urls (
                      id            BIGSERIAL PRIMARY KEY,
                      short_code    VARCHAR(16) NOT NULL UNIQUE,
                      original_url  TEXT NOT NULL,
                      access_count  BIGINT NOT NULL DEFAULT 0,
                      created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                      updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);