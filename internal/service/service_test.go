package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/emptypb"

	logger_lib "github.com/s21platform/logger-lib"

	"github.com/s21platform/materials-service/internal/config"
	"github.com/s21platform/materials-service/internal/model"
	"github.com/s21platform/materials-service/pkg/materials"
)

func TestServer_CreateMaterial(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockDBRepo(ctrl)
	mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

	ctx := context.Background()
	ctx = context.WithValue(ctx, config.KeyLogger, mockLogger)
	validUUID := uuid.New().String()
	ctxWithUUID := context.WithValue(ctx, config.KeyUUID, validUUID)
	s := New(mockRepo)

	t.Run("success", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("CreateMaterial")

		in := &materials.CreateMaterialIn{
			Title:           "New Material",
			CoverImageUrl:   "http://example.com/image.jpg",
			Description:     "Some description",
			Content:         "Some content",
			ReadTimeMinutes: 10,
		}

		expectedUUID := uuid.New().String()

		mockRepo.EXPECT().
			CreateMaterial(ctxWithUUID, validUUID, gomock.Any()).
			Return(expectedUUID, nil)

		out, err := s.CreateMaterial(ctxWithUUID, in)

		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.Equal(t, expectedUUID, out.Uuid)
	})

	t.Run("empty_title", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("CreateMaterial")
		mockLogger.EXPECT().Error("title is required")

		in := &materials.CreateMaterialIn{Title: ""}

		out, err := s.CreateMaterial(ctxWithUUID, in)

		assert.Nil(t, out)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "title is required")
	})

	t.Run("missing_owner_uuid", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("CreateMaterial")
		mockLogger.EXPECT().Error("uuid is required")

		in := &materials.CreateMaterialIn{Title: "Valid title"}

		out, err := s.CreateMaterial(ctx, in)

		assert.Nil(t, out)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "uuid is required")
	})

	t.Run("db_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("CreateMaterial")
		mockLogger.EXPECT().Error(gomock.Any())

		in := &materials.CreateMaterialIn{
			Title: "New Material",
		}

		mockRepo.EXPECT().
			CreateMaterial(ctxWithUUID, validUUID, gomock.Any()).
			Return("", fmt.Errorf("db failure"))

		out, err := s.CreateMaterial(ctxWithUUID, in)

		assert.Nil(t, out)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create material: db failure")
	})
}

func TestServer_GetAllMaterials(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockDBRepo(ctrl)
	mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

	createdAt := time.Now()
	editedAt := time.Now()
	publishedAt := time.Now()
	archivedAt := time.Now()
	deletedAt := time.Now()

	content1 := "Content for real 1"
	content2 := "Content for real 2"

	ctx := context.Background()
	ctx = context.WithValue(ctx, config.KeyLogger, mockLogger)

	s := New(mockRepo)

	t.Run("success", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("GetAllMaterials")

		expectedMaterials := model.MaterialList{
			{
				UUID:            uuid.New().String(),
				Title:           "Test Material one",
				Content:         &content1,
				OwnerUUID:       uuid.New().String(),
				CoverImageURL:   "http://example.com/cover1.jpg",
				Description:     "Description for real 1",
				ReadTimeMinutes: 5,
				Status:          "published",
				CreatedAt:       createdAt,
				EditedAt:        &editedAt,
				PublishedAt:     &publishedAt,
				ArchivedAt:      &archivedAt,
				DeletedAt:       &deletedAt,
				LikesCount:      10,
			},
			{
				UUID:            uuid.New().String(),
				Title:           "Test Material two",
				Content:         &content2,
				OwnerUUID:       uuid.New().String(),
				CoverImageURL:   "http://example.com/cover2.jpg",
				Description:     "Description for real 2",
				ReadTimeMinutes: 8,
				Status:          "draft",
				CreatedAt:       createdAt,
				EditedAt:        &editedAt,
				PublishedAt:     &publishedAt,
				ArchivedAt:      &archivedAt,
				DeletedAt:       &deletedAt,
				LikesCount:      20,
			},
		}

		mockRepo.EXPECT().GetAllMaterials(ctx).Return(&expectedMaterials, nil)

		materialsOut, err := s.GetAllMaterials(ctx, &emptypb.Empty{})

		assert.NoError(t, err)
		assert.NotNil(t, materialsOut)
		assert.Len(t, materialsOut.MaterialList, len(expectedMaterials))
		expectedMaterialList := expectedMaterials.ListFromDTO()
		assert.ElementsMatch(t, expectedMaterialList, materialsOut.MaterialList)
	})

	t.Run("DB_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("GetAllMaterials")
		mockLogger.EXPECT().Error(gomock.Any())

		mockRepo.EXPECT().GetAllMaterials(ctx).Return(nil, fmt.Errorf("failed to get all materials"))

		_, err := s.GetAllMaterials(ctx, &emptypb.Empty{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get all materials")
	})

	t.Run("empty_list", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("GetAllMaterials")

		emptyMaterials := model.MaterialList{}
		mockRepo.EXPECT().GetAllMaterials(ctx).Return(&emptyMaterials, nil)

		materialsOut, err := s.GetAllMaterials(ctx, &emptypb.Empty{})

		assert.NoError(t, err)
		assert.NotNil(t, materialsOut)
		assert.Empty(t, materialsOut.MaterialList)
	})
}
