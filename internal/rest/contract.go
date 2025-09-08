//go:generate mockgen -destination=mock_contract_test.go -package=${GOPACKAGE} -source=contract.go
package rest

import (
	"context"

	"github.com/s21platform/materials-service/internal/model"
)

type DBRepo interface {
	SaveDraftMaterial(ctx context.Context, ownerUUID string, material *model.SaveDraftMaterial) (string, error)
}
