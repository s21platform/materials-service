package service

import (
	"context"

	"github.com/s21platform/materials-service/pkg/materials"
)

type DBRepo interface {
	CreateMaterial(ctx context.Context, ownerUUID string, in *materials.CreateMaterialIn) (string, error)
}
