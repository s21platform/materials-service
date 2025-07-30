package service

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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

	return model.FromDTO(material), nil
}

func (s *Service) EditMaterial(ctx context.Context, in *materials.EditMaterialIn) (*materials.EditMaterialOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("EditMaterial")

	if strings.TrimSpace(in.Title) == "" {
		logger.Error("failed to get title")
		return nil, status.Error(codes.InvalidArgument, "failed to get title")
	}

	ownerUUID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok || ownerUUID == "" {
		logger.Error("failed to get uuid")
		return nil, status.Error(codes.Unauthenticated, "failed to get uuid")
	}

	materialOwnerUUID, err := s.repository.GetOwnerUUID(ctx, in.Uuid)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get owner uuid: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to get owner uuid: %v", err)
	}

	if materialOwnerUUID != ownerUUID {
		logger.Error("failed to edit: user is not owner")
		return nil, status.Errorf(codes.PermissionDenied, "failed to edit: user is not owner")
	}

	newMaterialData := &model.EditMaterial{}
	newMaterialData.ToDTO(in)

	err = s.repository.EditMaterial(ctx, newMaterialData)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to edit material: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to edit material: %v", err)
	}

	editedMaterial, err := s.repository.GetMaterial(ctx, in.Uuid)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get edited material: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to get edited material: %v", err)
	}

	return &materials.EditMaterialOut{
		Material: (*materials.Material)(model.FromDTO(editedMaterial)),
	}, nil
}
