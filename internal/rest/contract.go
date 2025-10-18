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
	CheckLike(ctx context.Context, materialUUID string, userUUID string) (bool, error)
	AddLike(ctx context.Context, materialUUID string, userUUID string) error
	RemoveLike(ctx context.Context, materialUUID string, userUUID string) error
	GetLikesCount(ctx context.Context, materialUUID string) (int32, error)
	UpdateLikesCount(ctx context.Context, materialUUID string, likesCount int32) error
	WithTx(ctx context.Context, cb func(ctx context.Context) error) (err error)
	EditMaterial(ctx context.Context, material *model.EditMaterial) (*model.Material, error)
	GetAllMaterials(ctx context.Context, offset, limit int) (*model.MaterialList, error)
	GetMaterial(ctx context.Context, materialUUID string) (*model.Material, error)
}

type KafkaProducer interface {
	ProduceMessage(ctx context.Context, message interface{}, key interface{}) error
}
