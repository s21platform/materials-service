//go:generate mockgen -destination=mock_contract_test.go -package=${GOPACKAGE} -source=contract.go
package rest

import (
	"context"

	"github.com/s21platform/materials-service/internal/model"
)

type DBRepo interface {
	SaveDraftMaterial(ctx context.Context, ownerUUID string, material *model.SaveDraftMaterial) (string, error)
	GetMaterialOwnerUUID(ctx context.Context, materialUUID string) (string, error)
	MaterialExists(ctx context.Context, materialUUID string) (bool, error)
	PublishMaterial(ctx context.Context, materialUUID string) (*model.Material, error)
}
