package model

import material_api "github.com/s21platform/materials-service/pkg/materials"

type CreateMaterial struct {
	Title           string `db:"title"`
	CoverImageURL   string `db:"cover_image_url"`
	Description     string `db:"description"`
	Content         string `db:"content"`
	ReadTimeMinutes int32  `db:"read_time_minutes"`
}

func (e *CreateMaterial) ToDTO(in *material_api.CreateMaterialIn) {
	e.Title = in.Title
	e.CoverImageURL = in.CoverImageUrl
	e.Description = in.Description
	e.Content = in.Content
	e.ReadTimeMinutes = int32(in.ReadTimeMinutes)
}
