-- +goose Up
CREATE TABLE IF NOT EXISTS materials
(
    uuid            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_uuid      UUID NOT NULL,
    title           TEXT NOT NULL,
    cover_image_uri TEXT,
    description     TEXT,
    content         TEXT,
    read_time_minutes INTEGER,
    status          TEXT NOT NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    edited_at       TIMESTAMP,
    published_at    TIMESTAMP,
    archived_at     TIMESTAMP,
    deleted_at      TIMESTAMP,
    likes_count     INTEGER DEFAULT 0,
    FOREIGN KEY (owner_uuid) REFERENCES users (uuid)
    );

-- +goose Down
DROP TABLE IF EXISTS materials;