//go:generate mockgen -destination=mock_contract_test.go -package=${GOPACKAGE} -source=contract.go
package service

import (
	"context"

	"github.com/s21platform/materials-service/internal/model"
)

type DBRepo interface {
	SaveDraftMaterial(ctx context.Context, ownerUUID string, material *model.SaveDraftMaterial) (string, error)
	GetMaterial(ctx context.Context, uuid string) (*model.Material, error)
	GetAllMaterials(ctx context.Context) (*model.MaterialList, error)
	EditMaterial(ctx context.Context, material *model.EditMaterial) (*model.Material, error)
	GetMaterialOwnerUUID(ctx context.Context, uuid string) (string, error)
	PublishMaterial(ctx context.Context, uuid string) (*model.Material, error)
}
