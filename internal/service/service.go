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

func (s *Service) SaveDraftMaterial(ctx context.Context, in *materials.SaveDraftMaterialIn) (*materials.SaveDraftMaterialOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("SaveDraftMaterial")

	if strings.TrimSpace(in.Title) == "" {
		logger.Error("title is required")
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	ownerUUID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok || ownerUUID == "" {
		logger.Error("uuid is required")
		return nil, status.Error(codes.Unauthenticated, "uuid is required")
	}

	newMaterialData := &model.SaveDraftMaterial{}
	newMaterialData.ToDTO(in)
	materialUUID, err := s.repository.SaveDraftMaterial(ctx, ownerUUID, newMaterialData)

	if err != nil {
		logger.Error(fmt.Sprintf("failed to save draft material: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to save draft material: %v", err)
	}

	return &materials.SaveDraftMaterialOut{
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

func (s *Service) PublishMaterial(ctx context.Context, in *materials.PublishMaterialIn) (*materials.PublishMaterialOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("PublishMaterial")

	if in.Uuid == "" {
		logger.Error("material uuid is required")
		return nil, status.Error(codes.InvalidArgument, "material uuid is required")
	}

	userUUID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok || userUUID == "" {
		logger.Error("user uuid is required")
		return nil, status.Error(codes.Unauthenticated, "user uuid is required")
	}

	materialOwnerUUID, err := s.repository.GetMaterialOwnerUUID(ctx, in.Uuid)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get owner uuid: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to get owner uuid: %v", err)
	}

	if materialOwnerUUID != userUUID {
		logger.Error("failed to publish: user is not owner")
		return nil, status.Errorf(codes.PermissionDenied, "failed to publish: user is not owner")
	}

	exists, err := s.repository.MaterialExists(ctx, in.Uuid)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to check material existence: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to check material existence: %v", err)
	}

	if !exists {
		logger.Error("material does not exist")
		return nil, status.Errorf(codes.FailedPrecondition, "material does not exist")
	}

	publishedMaterial, err := s.repository.PublishMaterial(ctx, in.Uuid)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to publish material: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to publish material: %v", err)
	}

	return &materials.PublishMaterialOut{
		Material: publishedMaterial.FromDTO(),
	}, nil
}
