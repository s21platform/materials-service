package service

import (
	"context"

	"github.com/s21platform/materials-service/model"
)

type DBRepo interface {
	CreateMaterial(ctx context.Context, ownerUUID string, material *model.CreateMaterial) (string, error)
}
