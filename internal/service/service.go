package service

import (
	"context"
	"fmt"
	logger_lib "github.com/s21platform/logger-lib"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/s21platform/materials-service/internal/config"
	_ "github.com/s21platform/materials-service/internal/repository/postgres"
	"github.com/s21platform/materials-service/pkg/materials"
)

type MaterialService struct {
	repository DBRepo
	materials.UnimplementedMaterialsServiceServer
}

type DBRepo interface {
	GetMaterial(ctx context.Context, uuid string) (*materials.GetMaterialOut, error)
}

func New(repo DBRepo) *MaterialService {
	return &MaterialService{
		repository: repo,
	}
}

func (s *MaterialService) GetMaterial(ctx context.Context, in *materials.GetMaterialIn) (*materials.GetMaterialOut, error) {
	logger := logger_lib.FromContext(ctx, config.KeyLogger)
	logger.AddFuncName("GetMaterial")

	_, ok := ctx.Value(config.KeyUUID).(string)
	if !ok {
		logger.Error("failed to find userUUID")
		return nil, status.Error(codes.Internal, "failed to find userUUID")
	}

	materialResponse, err := s.repository.GetMaterial(ctx, in.Uuid)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to find material: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to find material: %v", err)
	}

	return materialResponse, nil
}
