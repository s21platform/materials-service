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

func (s *Service) GetAllMaterials(ctx context.Context, _ *emptypb.Empty) (*materials.GetAllMaterialsOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("GetAllMaterials")

	uuid, ok := ctx.Value(config.KeyUUID).(string)
	if !ok {
		logger.Error("uuid is required")
		return nil, status.Error(codes.InvalidArgument, "uuid is required")
	}

	material, err := s.repository.GetAllMaterials(ctx, uuid)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get all materials: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to get all materials: %v", err)
	}

	return &materials.GetAllMaterialsOut{
		MaterialList: material.FromDTO(),
	}, nil
}
