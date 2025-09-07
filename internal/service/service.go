package service

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	logger_lib "github.com/s21platform/logger-lib"

	"github.com/s21platform/materials-service/internal/config"
	"github.com/s21platform/materials-service/internal/model"
	"github.com/s21platform/materials-service/internal/pkg/tx"
	"github.com/s21platform/materials-service/pkg/materials"
)

type Service struct {
	materials.UnimplementedMaterialsServiceServer
	repository DBRepo
}

func New(repo DBRepo) *Service {
	return &Service{
		repository: repo,
	}
}

func (s *Service) CreateMaterial(ctx context.Context, in *materials.CreateMaterialIn) (*materials.CreateMaterialOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("CreateMaterial")

	if strings.TrimSpace(in.Title) == "" {
		logger.Error("title is required")
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	ownerUUID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok || ownerUUID == "" {
		logger.Error("uuid is required")
		return nil, status.Error(codes.Unauthenticated, "uuid is required")
	}
	newMaterialData := &model.CreateMaterial{}
	newMaterialData.ToDTO(in)
	materialUUID, err := s.repository.CreateMaterial(ctx, ownerUUID, newMaterialData)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create material: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to create material: %v", err)
	}

	return &materials.CreateMaterialOut{
		Uuid: materialUUID,
	}, nil
}

func (s *Service) GetMaterial(ctx context.Context, in *materials.GetMaterialIn) (*materials.GetMaterialOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("GetMaterial")

	material, err := s.repository.GetMaterial(ctx, in.Uuid)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get material: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to get material: %v", err)
	}

	return &materials.GetMaterialOut{
		Material: material.FromDTO(),
	}, nil
}

func (s *Service) EditMaterial(ctx context.Context, in *materials.EditMaterialIn) (*materials.EditMaterialOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("EditMaterial")

	if strings.TrimSpace(in.Title) == "" {
		logger.Error("title is required")
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	userUUID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok || userUUID == "" {
		logger.Error("uuid is required")
		return nil, status.Error(codes.Unauthenticated, "uuid is required")
	}

	materialOwnerUUID, err := s.repository.GetMaterialOwnerUUID(ctx, in.Uuid)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get owner uuid: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to get owner uuid: %v", err)
	}

	if materialOwnerUUID != userUUID {
		logger.Error("failed to edit: user is not owner")
		return nil, status.Errorf(codes.PermissionDenied, "failed to edit: user is not owner")
	}

	updatedMaterial := &model.EditMaterial{}
	updatedMaterial.ToDTO(in)

	editedMaterial, err := s.repository.EditMaterial(ctx, updatedMaterial)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to edit material: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to edit material: %v", err)
	}

	return &materials.EditMaterialOut{
		Material: editedMaterial.FromDTO(),
	}, nil
}

func (s *Service) GetAllMaterials(ctx context.Context, _ *emptypb.Empty) (*materials.GetAllMaterialsOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("GetAllMaterials")

	materialsList, err := s.repository.GetAllMaterials(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get all materials: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to get all materials: %v", err)
	}

	return &materials.GetAllMaterialsOut{
		MaterialList: materialsList.ListFromDTO(),
	}, nil
}

func (s *Service) ToggleLike(ctx context.Context, in *materials.ToggleLikeIn) (*materials.ToggleLikeOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("ToggleLike")

	userUUID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok || userUUID == "" {
		logger.Error("uuid is required")
		return nil, status.Error(codes.Unauthenticated, "uuid is required")
	}

	var isLiked bool
	var likesCount int32
	err := tx.TxExecute(ctx, func(ctx context.Context) error {
		var err error
		isLiked, err = s.repository.CheckLike(ctx, in.MaterialUuid, userUUID)
		if err != nil {
			return fmt.Errorf("failed to check like: %v", err)
		}

		if isLiked {
			_, err = s.repository.RemoveLike(ctx, in.MaterialUuid, userUUID)
			if err != nil {
				return fmt.Errorf("failed to remove like: %v", err)
			}
		} else {
			_, err = s.repository.AddLike(ctx, in.MaterialUuid, userUUID)
			if err != nil {
				return fmt.Errorf("failed to add like: %v", err)
			}
		}

		likesCount, err = s.repository.GetLikesCount(ctx, in.MaterialUuid)
		if err != nil {
			return fmt.Errorf("failed to get likes count: %v", err)
		}

		_, err = s.repository.UpdateLikesCount(ctx, in.MaterialUuid, likesCount)
		if err != nil {
			return fmt.Errorf("failed to update likes count: %v", err)
		}

		return nil
	})
	if err != nil {
		logger.Error(fmt.Sprintf("transaction failed: %v", err))
		return nil, status.Errorf(codes.Internal, "transaction failed: %v", err)
	}

	return &materials.ToggleLikeOut{
		IsLiked:    !isLiked,
		LikesCount: likesCount,
	}, nil
}
