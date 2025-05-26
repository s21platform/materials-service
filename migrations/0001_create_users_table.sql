-- +goose Up
CREATE TABLE IF NOT EXISTS users
(
    uuid          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nickname      TEXT NOT NULL,
    avatar_link   TEXT NOT NULL,
    name          TEXT,
    surname       TEXT,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

-- +goose Down
DROP TABLE IF EXISTS users;