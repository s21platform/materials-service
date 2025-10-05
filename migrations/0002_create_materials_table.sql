-- +goose Up
CREATE TYPE material_status AS ENUM ('draft', 'published', 'archived');

CREATE TABLE IF NOT EXISTS materials
(
    uuid            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_uuid      UUID NOT NULL,
    title           TEXT NOT NULL,
    cover_image_url TEXT NOT NULL,
    description     TEXT NOT NULL,
    content         TEXT NOT NULL,
    read_time_minutes INTEGER,
    status          material_status NOT NULL DEFAULT 'draft',
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
DROP TYPE IF EXISTS material_status;