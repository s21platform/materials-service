package model

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/s21platform/materials-service/pkg/materials"
)

type MaterialList []Material

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
	if m.EditedAt != nil {
		protoMaterial.EditedAt = timestamppb.New(*m.EditedAt)
	}
	if m.PublishedAt != nil {
		protoMaterial.PublishedAt = timestamppb.New(*m.PublishedAt)
	}
	if m.ArchivedAt != nil {
		protoMaterial.ArchivedAt = timestamppb.New(*m.ArchivedAt)
	}
	if m.DeletedAt != nil {
		protoMaterial.DeletedAt = timestamppb.New(*m.DeletedAt)
	}

	return protoMaterial
}

func (a *MaterialList) ListFromDTO() []*materials.Material {
	result := make([]*materials.Material, 0, len(*a))

	for _, material := range *a {
		m := &materials.Material{
			Uuid:            material.UUID,
			OwnerUuid:       material.OwnerUUID,
			Title:           material.Title,
			CoverImageUrl:   material.CoverImageURL,
			Description:     material.Description,
			ReadTimeMinutes: material.ReadTimeMinutes,
			Status:          material.Status,
			LikesCount:      material.LikesCount,
		}

		if material.Content != nil {
			m.Content = *material.Content
		}
		if material.EditedAt != nil {
			m.EditedAt = timestamppb.New(*material.EditedAt)
		}
		if material.PublishedAt != nil {
			m.PublishedAt = timestamppb.New(*material.PublishedAt)
		}
		if material.ArchivedAt != nil {
			m.ArchivedAt = timestamppb.New(*material.ArchivedAt)
		}
		if material.DeletedAt != nil {
			m.DeletedAt = timestamppb.New(*material.DeletedAt)
		}

		result = append(result, m)
	}

	return result
}
