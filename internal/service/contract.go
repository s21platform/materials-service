//go:generate mockgen -destination=mock_contract_test.go -package=${GOPACKAGE} -source=contract.go
package service

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/s21platform/materials-service/internal/model"
)

type DBRepo interface {
	Conn() *sqlx.DB
	CreateMaterial(ctx context.Context, ownerUUID string, material *model.CreateMaterial) (string, error)
	GetMaterial(ctx context.Context, uuid string) (*model.Material, error)
	GetAllMaterials(ctx context.Context) (*model.MaterialList, error)
	EditMaterial(ctx context.Context, material *model.EditMaterial) (*model.Material, error)
	GetMaterialOwnerUUID(ctx context.Context, uuid string) (string, error)
	CheckLike(ctx context.Context, materialUUID string, userUUID string, tx *sqlx.Tx) (bool, error)
	AddLike(ctx context.Context, materialUUID string, userUUID string, tx *sqlx.Tx) (bool, error)
	RemoveLike(ctx context.Context, materialUUID string, userUUID string, tx *sqlx.Tx) (bool, error)
	GetLikesCount(ctx context.Context, materialUUID string, tx *sqlx.Tx) (int32, error)
	UpdateLikesCount(ctx context.Context, materialUUID string, likesCount int32, tx *sqlx.Tx) (int32, error)
}
