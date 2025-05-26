-- +goose Up
CREATE TABLE IF NOT EXISTS material_likes
(
    uuid          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    material_uuid UUID NOT NULL,
    user_uuid     UUID NOT NULL,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (material_uuid) REFERENCES materials (uuid),
    FOREIGN KEY (user_uuid) REFERENCES users (uuid),
    CONSTRAINT unique_material_user UNIQUE (material_uuid, user_uuid)
    );

-- +goose Down
DROP TABLE IF EXISTS material_likes;