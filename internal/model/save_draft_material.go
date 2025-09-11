package model

import "github.com/s21platform/materials-service/pkg/materials"

type SaveDraftMaterial struct {
	Title           string `db:"title"`
	CoverImageURL   string `db:"cover_image_url"`
	Description     string `db:"description"`
	Content         string `db:"content"`
	ReadTimeMinutes int32  `db:"read_time_minutes"`
}

func (e *SaveDraftMaterial) ToDTO(in *materials.SaveDraftMaterialIn) {
	e.Title = in.Title
	e.CoverImageURL = in.CoverImageUrl
	e.Description = in.Description
	e.Content = in.Content
	e.ReadTimeMinutes = in.ReadTimeMinutes
}
