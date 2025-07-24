package model

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/s21platform/materials-service/pkg/materials"
)

type Material struct {
	UUID            string     `db:"uuid"`
	OwnerUUID       string     `db:"owner_uuid"`
	Title           string     `db:"title"`
	CoverImageURL   *string    `db:"cover_image_url"`
	Description     *string    `db:"description"`
	Content         *string    `db:"content"`
	ReadTimeMinutes *int32     `db:"read_time_minutes"`
	Status          string     `db:"status"`
	CreatedAt       time.Time  `db:"created_at"`
	EditedAt        *time.Time `db:"edited_at"`
	PublishedAt     *time.Time `db:"published_at"`
	ArchivedAt      *time.Time `db:"archived_at"`
	DeletedAt       *time.Time `db:"deleted_at"`
	LikesCount      int32      `db:"likes_count"`
}

func FromDTO(material *Material) *materials.GetMaterialOut {
	protoMaterial := &materials.GetMaterialOut{
		Uuid:       material.UUID,
		OwnerUuid:  material.OwnerUUID,
		Title:      material.Title,
		LikesCount: material.LikesCount,
	}

	if material.CoverImageURL != nil {
		protoMaterial.CoverImageUrl = *material.CoverImageURL
	}
	if material.Description != nil {
		protoMaterial.Description = *material.Description
	}
	if material.Content != nil {
		protoMaterial.Content = *material.Content
	}
	if material.ReadTimeMinutes != nil {
		protoMaterial.ReadTimeMinutes = *material.ReadTimeMinutes
	}
	if !material.CreatedAt.IsZero() {
		protoMaterial.CreatedAt = timestamppb.New(material.CreatedAt)
	}
	if material.EditedAt != nil && !material.EditedAt.IsZero() {
		protoMaterial.EditedAt = timestamppb.New(*material.EditedAt)
	}
	if material.PublishedAt != nil && !material.PublishedAt.IsZero() {
		protoMaterial.PublishedAt = timestamppb.New(*material.PublishedAt)
	}
	if material.ArchivedAt != nil && !material.ArchivedAt.IsZero() {
		protoMaterial.ArchivedAt = timestamppb.New(*material.ArchivedAt)
	}
	if material.DeletedAt != nil && !material.DeletedAt.IsZero() {
		protoMaterial.DeletedAt = timestamppb.New(*material.DeletedAt)
	}

	return protoMaterial
}
