package service

import (
	"context"
	"fmt"
	"strings"

	logger_lib "github.com/s21platform/logger-lib"
	"github.com/s21platform/materials-service/internal/config"
	"github.com/s21platform/materials-service/internal/model"
	"github.com/s21platform/materials-service/internal/pkg/tx"
	"github.com/s21platform/materials-service/pkg/materials"
	
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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

func (s *Service) SaveDraftMaterial(ctx context.Context, in *materials.SaveDraftMaterialIn) (*materials.SaveDraftMaterialOut, error) {
	if strings.TrimSpace(in.Title) == "" {
		logger_lib.Error(ctx, "title is required")
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	ownerUUID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok || ownerUUID == "" {
		logger_lib.Error(ctx, "uuid is required")
		return nil, status.Error(codes.Unauthenticated, "uuid is required")
	}

	ctx = logger_lib.WithUserUuid(ctx, ownerUUID)

	newMaterialData := &model.SaveDraftMaterial{}
	newMaterialData.ToDTO(in)
	materialUUID, err := s.repository.SaveDraftMaterial(ctx, ownerUUID, newMaterialData)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to save draft material")
		return nil, status.Errorf(codes.Internal, "failed to save draft material: %v", err)
	}

	return &materials.SaveDraftMaterialOut{
		Uuid: materialUUID,
	}, nil
}

func (s *Service) GetMaterial(ctx context.Context, in *materials.GetMaterialIn) (*materials.GetMaterialOut, error) {
	material, err := s.repository.GetMaterial(ctx, in.Uuid)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to get material")
		return nil, status.Errorf(codes.Internal, "failed to get material: %v", err)
	}

	return &materials.GetMaterialOut{
		Material: material.FromDTO(),
	}, nil
}

func (s *Service) EditMaterial(ctx context.Context, in *materials.EditMaterialIn) (*materials.EditMaterialOut, error) {
	if strings.TrimSpace(in.Title) == "" {
		logger_lib.Error(ctx, "title is required")
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	userUUID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok || userUUID == "" {
		logger_lib.Error(ctx, "uuid is required")
		return nil, status.Error(codes.Unauthenticated, "uuid is required")
	}

	ctx = logger_lib.WithUserUuid(ctx, userUUID)

	materialOwnerUUID, err := s.repository.GetMaterialOwnerUUID(ctx, in.Uuid)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to get owner uuid")
		return nil, status.Errorf(codes.Internal, "failed to get owner uuid: %v", err)
	}

	if materialOwnerUUID != userUUID {
		logger_lib.Error(ctx, "failed to edit: user is not owner")
		return nil, status.Errorf(codes.PermissionDenied, "failed to edit: user is not owner")
	}

	updatedMaterial := &model.EditMaterial{}
	updatedMaterial.ToDTO(in)

	editedMaterial, err := s.repository.EditMaterial(ctx, updatedMaterial)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to edit material")
		return nil, status.Errorf(codes.Internal, "failed to edit material: %v", err)
	}

	return &materials.EditMaterialOut{
		Material: editedMaterial.FromDTO(),
	}, nil
}

func (s *Service) GetAllMaterials(ctx context.Context, _ *emptypb.Empty) (*materials.GetAllMaterialsOut, error) {
	materialsList, err := s.repository.GetAllMaterials(ctx)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to get all materials")
		return nil, status.Errorf(codes.Internal, "failed to get all materials: %v", err)
	}

	return &materials.GetAllMaterialsOut{
		MaterialList: materialsList.ListFromDTO(),
	}, nil
}

func (s *Service) DeleteMaterial(ctx context.Context, in *materials.DeleteMaterialIn) (*emptypb.Empty, error) {
	userUUID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok || userUUID == "" {
		logger_lib.Error(ctx, "uuid is required")
		return nil, status.Error(codes.Unauthenticated, "uuid is required")
	}

	ctx = logger_lib.WithUserUuid(ctx, userUUID)

	materialOwnerUUID, err := s.repository.GetMaterialOwnerUUID(ctx, in.Uuid)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to get owner uuid")
		return nil, status.Errorf(codes.Internal, "failed to get owner uuid: %v", err)
	}

	if materialOwnerUUID != userUUID {
		logger_lib.Error(ctx, "failed to delete: user is not owner")
		return nil, status.Errorf(codes.PermissionDenied, "failed to delete: user is not owner")
	}

	rowsAffected, err := s.repository.DeleteMaterial(ctx, in.Uuid)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to delete material")
		return nil, status.Errorf(codes.Internal, "failed to delete material: %v", err)
	}

	if rowsAffected == 0 {
		logger_lib.Error(ctx, "failed to delete: material already deleted or not found")
		return nil, status.Errorf(codes.NotFound, "failed to delete: material already deleted or not found")
	}

	return nil, nil
}

func (s *Service) ArchivedMaterial(ctx context.Context, in *materials.ArchivedMaterialIn) (*emptypb.Empty, error) {
	userUUID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok || userUUID == "" {
		logger_lib.Error(ctx, "uuid is required")
		return nil, status.Error(codes.Unauthenticated, "uuid is required")
	}

	ctx = logger_lib.WithUserUuid(ctx, userUUID)

	materialOwnerUUID, err := s.repository.GetMaterialOwnerUUID(ctx, in.Uuid)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to get owner uuid")
		return nil, status.Errorf(codes.Internal, "failed to get owner uuid: %v", err)
	}

	if materialOwnerUUID != userUUID {
		logger_lib.Error(ctx, "failed to archive: user is not owner")
		return nil, status.Errorf(codes.PermissionDenied, "failed to archive: user is not owner")
	}

	rowsAffected, err := s.repository.ArchivedMaterial(ctx, in.Uuid)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to archive material")
		return nil, status.Errorf(codes.Internal, "failed to archive material: %v", err)
	}

	if rowsAffected == 0 {
		logger_lib.Error(ctx, "failed to archive material: material already archived or not found")
		return nil, status.Error(codes.NotFound, "failed to archive material: material already archived or not found")
	}

	return nil, nil
}

func (s *Service) PublishMaterial(ctx context.Context, in *materials.PublishMaterialIn) (*materials.PublishMaterialOut, error) {
	if in.Uuid == "" {
		logger_lib.Error(ctx, "material uuid is required")
		return nil, status.Error(codes.InvalidArgument, "material uuid is required")
	}

	userUUID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok || userUUID == "" {
		logger_lib.Error(ctx, "user uuid is required")
		return nil, status.Error(codes.Unauthenticated, "user uuid is required")
	}

	ctx = logger_lib.WithUserUuid(ctx, userUUID)

	materialOwnerUUID, err := s.repository.GetMaterialOwnerUUID(ctx, in.Uuid)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to get owner uuid")
		return nil, status.Errorf(codes.Internal, "failed to get owner uuid: %v", err)
	}

	if materialOwnerUUID != userUUID {
		logger_lib.Error(ctx, "failed to publish: user is not owner")
		return nil, status.Errorf(codes.PermissionDenied, "failed to publish: user is not owner")
	}

	exists, err := s.repository.MaterialExists(ctx, in.Uuid)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to check material existence")
		return nil, status.Errorf(codes.Internal, "failed to check material existence: %v", err)
	}

	if !exists {
		logger_lib.Error(ctx, "material does not exist")
		return nil, status.Errorf(codes.FailedPrecondition, "material does not exist")
	}

	publishedMaterial, err := s.repository.PublishMaterial(ctx, in.Uuid)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to publish material")
		return nil, status.Errorf(codes.Internal, "failed to publish material: %v", err)
	}

	return &materials.PublishMaterialOut{
		Material: publishedMaterial.FromDTO(),
	}, nil
}

func (s *Service) ToggleLike(ctx context.Context, in *materials.ToggleLikeIn) (*materials.ToggleLikeOut, error) {
	userUUID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok || userUUID == "" {
		logger_lib.Error(ctx, "uuid is required")
		return nil, status.Error(codes.Unauthenticated, "uuid is required")
	}

	ctx = logger_lib.WithUserUuid(ctx, userUUID)

	var isLiked bool
	var likesCount int32
	err := tx.TxExecute(ctx, func(ctx context.Context) error {
		var err error
		isLiked, err = s.repository.CheckLike(ctx, in.MaterialUuid, userUUID)
		if err != nil {
			return fmt.Errorf("failed to check like: %v", err)
		}

		if isLiked {
			err = s.repository.RemoveLike(ctx, in.MaterialUuid, userUUID)
			if err != nil {
				return fmt.Errorf("failed to remove like: %v", err)
			}
		} else {
			err = s.repository.AddLike(ctx, in.MaterialUuid, userUUID)
			if err != nil {
				return fmt.Errorf("failed to add like: %v", err)
			}
		}

		likesCount, err = s.repository.GetLikesCount(ctx, in.MaterialUuid)
		if err != nil {
			return fmt.Errorf("failed to get likes count: %v", err)
		}

		err = s.repository.UpdateLikesCount(ctx, in.MaterialUuid, likesCount)
		if err != nil {
			return fmt.Errorf("failed to update likes count: %v", err)
		}

		return nil
	})
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "transaction failed")
		return nil, status.Errorf(codes.Internal, "transaction failed: %v", err)
	}

	return &materials.ToggleLikeOut{
		IsLiked:    !isLiked,
		LikesCount: likesCount,
	}, nil
}
