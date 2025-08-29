//go:generate mockgen -destination=mock_contract_test.go -package=${GOPACKAGE} -source=contract.go
package service

import (
	"context"

	"github.com/s21platform/materials-service/internal/model"
)

type DBRepo interface {
	CreateMaterial(ctx context.Context, ownerUUID string, material *model.CreateMaterial) (string, error)
	GetMaterial(ctx context.Context, uuid string) (*model.Material, error)
	GetAllMaterials(ctx context.Context) (*model.MaterialList, error)
	EditMaterial(ctx context.Context, material *model.EditMaterial) (*model.Material, error)
	GetMaterialOwnerUUID(ctx context.Context, uuid string) (string, error)
	ToggleLike(ctx context.Context, materialUUID string, userUUID string) (bool, error)
	UpdateLikesNumber(ctx context.Context, materialUUID string) (int32, error)
}
