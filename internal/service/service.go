package service

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	logger_lib "github.com/s21platform/logger-lib"

	"github.com/s21platform/materials-service/internal/config"
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

	uuid, ok := ctx.Value(config.KeyUUID).(string)
	if !ok || uuid == "" {
		logger.Error("uuid is required in context")
		return nil, status.Error(codes.Unauthenticated, "user uuid is required")
	}

	materialUUID, err := s.repository.CreateMaterial(ctx, uuid, in)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create material: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to create material: %v", err)
	}

	return &materials.CreateMaterialOut{
		Uuid: materialUUID,
	}, nil
}
