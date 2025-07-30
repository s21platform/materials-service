package service

import (
	"context"

	"github.com/s21platform/materials-service/internal/model"
)

type DBRepo interface {
	CreateMaterial(ctx context.Context, ownerUUID string, material *model.CreateMaterial) (string, error)
	GetMaterial(ctx context.Context, uuid string) (*model.Material, error)
	EditMaterial(ctx context.Context, data *model.EditMaterial) error
	GetOwnerUUID(ctx context.Context, uuid string) (string, error)
}
