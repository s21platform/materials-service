package service

import (
	"context"

	"github.com/s21platform/materials-service/internal/model"
)

type DBRepo interface {
	GetMaterial(ctx context.Context, uuid string) (*model.Material, error)
}
