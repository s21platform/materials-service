package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	logger_lib "github.com/s21platform/logger-lib"

	"github.com/s21platform/materials-service/internal/config"
	"github.com/s21platform/materials-service/internal/model"
	"github.com/s21platform/materials-service/pkg/materials"
)

func TestServer_GetMaterial(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockDBRepo(ctrl)
	mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

	userUUID := uuid.New().String()
	materialUUID := uuid.New().String()
	title := "Material Title"
	editedAt := time.Now()
	publishedAt := time.Now()
	archivedAt := time.Now()
	deletedAt := time.Now()
	content := "material content"

	ctx := context.Background()
	ctx = context.WithValue(ctx, config.KeyLogger, mockLogger)

	s := New(mockRepo)

	t.Run("success", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("GetMaterial")

		in := &materials.GetMaterialIn{
			Uuid: materialUUID,
		}

		expectedMaterial := &model.Material{
			UUID:            materialUUID,
			OwnerUUID:       userUUID,
			Title:           title,
			CoverImageURL:   "test",
			Description:     "cool fr",
			Content:         &content,
			ReadTimeMinutes: 12,
			Status:          "approved",
			CreatedAt:       time.Now(),
			EditedAt:        &editedAt,
			PublishedAt:     &publishedAt,
			ArchivedAt:      &archivedAt,
			DeletedAt:       &deletedAt,
			LikesCount:      10,
		}

		mockRepo.EXPECT().GetMaterial(ctx, materialUUID).Return(expectedMaterial, nil)

		expectedOutput := &materials.GetMaterialOut{
			Material: expectedMaterial.FromDTO(),
		}

		out, err := s.GetMaterial(ctx, in)

		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.Equal(t, expectedOutput, out)
	})

	t.Run("get_material_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("GetMaterial")
		dbErr := fmt.Errorf("database error")
		mockLogger.EXPECT().Error(fmt.Sprintf("failed to get material: %v", dbErr))

		mockRepo.EXPECT().GetMaterial(ctx, materialUUID).Return(nil, dbErr)

		in := &materials.GetMaterialIn{Uuid: materialUUID}
		_, err := s.GetMaterial(ctx, in)

		assert.Error(t, err)
		sts := status.Convert(err)
		assert.Equal(t, codes.Internal, sts.Code())
		assert.Contains(t, sts.Message(), "failed to get material: database error")
	})
}

func TestServer_EditMaterial(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockDBRepo(ctrl)
	mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

	userUUID := uuid.New().String()
	materialUUID := uuid.New().String()
	newTitle := "New Title"
	editedAt := time.Now()
	publishedAt := time.Now()
	archivedAt := time.Now()
	deletedAt := time.Now()
	content := "updated content"

	ctx := context.Background()
	ctx = context.WithValue(ctx, config.KeyLogger, mockLogger)
	ctx = context.WithValue(ctx, config.KeyUUID, userUUID)

	s := New(mockRepo)

	t.Run("success", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("EditMaterial")

		mockRepo.EXPECT().GetMaterialOwnerUUID(ctx, materialUUID).Return(userUUID, nil)

		in := &materials.EditMaterialIn{
			Uuid:            materialUUID,
			Title:           newTitle,
			CoverImageUrl:   "test",
			Description:     "new feature",
			Content:         "super new",
			ReadTimeMinutes: 11,
		}

		expectedMaterial := &model.Material{
			UUID:            materialUUID,
			OwnerUUID:       userUUID,
			Title:           newTitle,
			CoverImageURL:   "test",
			Description:     "new feature",
			Content:         &content,
			ReadTimeMinutes: 11,
			Status:          "published",
			CreatedAt:       time.Now(),
			EditedAt:        &editedAt,
			PublishedAt:     &publishedAt,
			ArchivedAt:      &archivedAt,
			DeletedAt:       &deletedAt,
			LikesCount:      10,
		}

		mockRepo.EXPECT().EditMaterial(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, em *model.EditMaterial) (*model.Material, error) {
			return expectedMaterial, nil
		})

		expectedOutput := &materials.EditMaterialOut{
			Material: expectedMaterial.FromDTO(),
		}

		out, err := s.EditMaterial(ctx, in)

		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.Equal(t, expectedOutput, out)
	})

	t.Run("empty_title", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("EditMaterial")
		mockLogger.EXPECT().Error("title is required")

		in := &materials.EditMaterialIn{Uuid: materialUUID, Title: ""}
		_, err := s.EditMaterial(ctx, in)

		assert.Error(t, err)
		sts := status.Convert(err)
		assert.Equal(t, codes.InvalidArgument, sts.Code())
		assert.Equal(t, "title is required", sts.Message())
	})

	t.Run("title_with_only_spaces", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("EditMaterial")
		mockLogger.EXPECT().Error("title is required")

		in := &materials.EditMaterialIn{Uuid: materialUUID, Title: "   "}
		_, err := s.EditMaterial(ctx, in)

		assert.Error(t, err)
		sts := status.Convert(err)
		assert.Equal(t, codes.InvalidArgument, sts.Code())
		assert.Equal(t, "title is required", sts.Message())
	})

	t.Run("no_user_uuid", func(t *testing.T) {
		badCtx := context.Background()
		badCtx = context.WithValue(badCtx, config.KeyLogger, mockLogger)

		mockLogger.EXPECT().AddFuncName("EditMaterial")
		mockLogger.EXPECT().Error("uuid is required")

		in := &materials.EditMaterialIn{Uuid: materialUUID, Title: newTitle}
		_, err := s.EditMaterial(badCtx, in)

		assert.Error(t, err)
		sts := status.Convert(err)
		assert.Equal(t, codes.Unauthenticated, sts.Code())
		assert.Equal(t, "uuid is required", sts.Message())
	})

	t.Run("get_owner_uuid_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("EditMaterial")
		dbErr := fmt.Errorf("database error")
		mockLogger.EXPECT().Error(fmt.Sprintf("failed to get owner uuid: %v", dbErr))

		mockRepo.EXPECT().GetMaterialOwnerUUID(ctx, materialUUID).Return("", dbErr)

		in := &materials.EditMaterialIn{Uuid: materialUUID, Title: newTitle}
		_, err := s.EditMaterial(ctx, in)

		assert.Error(t, err)
		sts := status.Convert(err)
		assert.Equal(t, codes.Internal, sts.Code())
		assert.Contains(t, sts.Message(), "failed to get owner uuid: database error")
	})

	t.Run("user_not_owner", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("EditMaterial")
		mockLogger.EXPECT().Error("failed to edit: user is not owner")

		mockRepo.EXPECT().GetMaterialOwnerUUID(ctx, materialUUID).Return("other-uuid", nil)

		in := &materials.EditMaterialIn{Uuid: materialUUID, Title: newTitle}
		_, err := s.EditMaterial(ctx, in)

		assert.Error(t, err)
		sts := status.Convert(err)
		assert.Equal(t, codes.PermissionDenied, sts.Code())
		assert.Equal(t, "failed to edit: user is not owner", sts.Message())
	})

	t.Run("edit_material_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("EditMaterial")
		dbErr := fmt.Errorf("edit failed")
		mockLogger.EXPECT().Error(fmt.Sprintf("failed to edit material: %v", dbErr))

		mockRepo.EXPECT().GetMaterialOwnerUUID(ctx, materialUUID).Return(userUUID, nil)

		mockRepo.EXPECT().EditMaterial(ctx, gomock.Any()).Return(nil, dbErr)

		in := &materials.EditMaterialIn{Uuid: materialUUID, Title: newTitle}
		_, err := s.EditMaterial(ctx, in)

		assert.Error(t, err)
		sts := status.Convert(err)
		assert.Equal(t, codes.Internal, sts.Code())
		assert.Contains(t, sts.Message(), "failed to edit material: edit failed")
	})
}
