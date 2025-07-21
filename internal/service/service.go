package service

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	logger_lib "github.com/s21platform/logger-lib"

	"github.com/s21platform/materials-service/internal/config"
	"github.com/s21platform/materials-service/pkg/materials"
)

type DBRepo interface {
	CreateMaterial(ctx context.Context, in *materials.CreateMaterialIn) (*materials.CreateMaterialOut, error)
}

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
	if logger == nil {
		return nil, status.Error(codes.Internal, "logger not initialized in context")
	}
	logger.AddFuncName("CreateMaterial")
	if s.repository == nil {
		logger.Error("repository is not initialized")
		return nil, status.Error(codes.Internal, "internal server error: repository not initialized")
	}
	ownerUUID, ok := ctx.Value(config.KeyUUID).(string)
	if !ok {
		logger.Error("owner uuid is required")
		return nil, status.Error(codes.InvalidArgument, "owner uuid is required")
	}

	if in.OwnerUuid != ownerUUID {
		logger.Error("owner uuid mismatch")
		return nil, status.Error(codes.PermissionDenied, "owner uuid mismatch")
	}

	if in.Title == "" {
		logger.Error("title is required")
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	//materialUUID := uuid.New().String()

	out, err := s.repository.CreateMaterial(ctx, in)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create material: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to create material: %v", err)
	}

	return out, nil
}
