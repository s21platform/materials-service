package model

import (
	"time"

	"github.com/s21platform/materials-service/pkg/materials"
)

type Material struct {
	UUID            string     `db:"uuid"`
	OwnerUUID       string     `db:"owner_uuid"`
	Title           string     `db:"title"`
	CoverImageURL   string     `db:"cover_image_url"`
	Description     string     `db:"description"`
	Content         *string    `db:"content"`
	ReadTimeMinutes int32      `db:"read_time_minutes"`
	Status          string     `db:"status"`
	CreatedAt       time.Time  `db:"created_at"`
	EditedAt        *time.Time `db:"edited_at"`
	PublishedAt     *time.Time `db:"published_at"`
	ArchivedAt      *time.Time `db:"archived_at"`
	DeletedAt       *time.Time `db:"deleted_at"`
	LikesCount      int32      `db:"likes_count"`
}

func (m *Material) FromDTO() *materials.Material {
	protoMaterial := &materials.Material{
		Uuid:       m.UUID,
		OwnerUuid:  m.OwnerUUID,
		Title:      m.Title,
		LikesCount: m.LikesCount,
	}

	if m.Content != nil {
		protoMaterial.Content = *m.Content
	}

	return protoMaterial
}
