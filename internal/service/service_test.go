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
	"google.golang.org/protobuf/types/known/emptypb"

	logger_lib "github.com/s21platform/logger-lib"

	"github.com/s21platform/materials-service/internal/config"
	"github.com/s21platform/materials-service/internal/model"
	"github.com/s21platform/materials-service/internal/pkg/tx"
	"github.com/s21platform/materials-service/pkg/materials"
)

func TestServer_SaveDraftMaterial(t *testing.T) {
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
		mockLogger.EXPECT().AddFuncName("SaveDraftMaterial")

		in := &materials.SaveDraftMaterialIn{
			Title:           "New Material",
			CoverImageUrl:   "http://example.com/image.jpg",
			Description:     "Some description",
			Content:         "Some content",
			ReadTimeMinutes: 10,
		}

		expectedUUID := uuid.New().String()

		mockRepo.EXPECT().
			SaveDraftMaterial(ctxWithUUID, validUUID, gomock.Any()).
			Return(expectedUUID, nil)

		out, err := s.SaveDraftMaterial(ctxWithUUID, in)

		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.Equal(t, expectedUUID, out.Uuid)
	})

	t.Run("empty_title", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("SaveDraftMaterial")
		mockLogger.EXPECT().Error("title is required")

		in := &materials.SaveDraftMaterialIn{Title: ""}

		out, err := s.SaveDraftMaterial(ctxWithUUID, in)

		assert.Nil(t, out)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "title is required")
	})

	t.Run("missing_owner_uuid", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("SaveDraftMaterial")
		mockLogger.EXPECT().Error("uuid is required")

		in := &materials.SaveDraftMaterialIn{Title: "Valid title"}

		out, err := s.SaveDraftMaterial(ctx, in)

		assert.Nil(t, out)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "uuid is required")
	})

	t.Run("db_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("SaveDraftMaterial")
		mockLogger.EXPECT().Error(gomock.Any())

		in := &materials.SaveDraftMaterialIn{
			Title: "New Material",
		}

		mockRepo.EXPECT().
			SaveDraftMaterial(ctxWithUUID, validUUID, gomock.Any()).
			Return("", fmt.Errorf("db failure"))

		out, err := s.SaveDraftMaterial(ctxWithUUID, in)

		assert.Nil(t, out)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save draft material: db failure")
	})
}

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

		mockRepo.EXPECT().EditMaterial(ctx, gomock.Any()).Return(expectedMaterial, nil)

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

func TestServer_DeleteMaterial(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockDBRepo(ctrl)
	mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

	userUUID := uuid.New().String()
	materialUUID := uuid.New().String()

	ctx := context.Background()
	ctx = context.WithValue(ctx, config.KeyLogger, mockLogger)
	ctx = context.WithValue(ctx, config.KeyUUID, userUUID)

	s := New(mockRepo)

	t.Run("success", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("DeleteMaterial")
		mockRepo.EXPECT().GetMaterialOwnerUUID(ctx, materialUUID).Return(userUUID, nil)
		mockRepo.EXPECT().DeleteMaterial(ctx, materialUUID).Return(int64(1), nil)

		in := &materials.DeleteMaterialIn{Uuid: materialUUID}
		out, err := s.DeleteMaterial(ctx, in)

		assert.NoError(t, err)
		assert.Nil(t, out)
	})

	t.Run("no_user_uuid", func(t *testing.T) {
		badCtx := context.Background()
		badCtx = context.WithValue(badCtx, config.KeyLogger, mockLogger)
		mockLogger.EXPECT().AddFuncName("DeleteMaterial")
		mockLogger.EXPECT().Error("uuid is required")

		in := &materials.DeleteMaterialIn{Uuid: materialUUID}
		_, err := s.DeleteMaterial(badCtx, in)

		assert.Error(t, err)
		sts := status.Convert(err)
		assert.Equal(t, codes.Unauthenticated, sts.Code())
		assert.Equal(t, "uuid is required", sts.Message())
	})

	t.Run("get_owner_uuid_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("DeleteMaterial")

		dbErr := fmt.Errorf("database error")
		mockLogger.EXPECT().Error(fmt.Sprintf("failed to get owner uuid: %v", dbErr))

		mockRepo.EXPECT().GetMaterialOwnerUUID(ctx, materialUUID).Return("", dbErr)

		in := &materials.DeleteMaterialIn{Uuid: materialUUID}
		_, err := s.DeleteMaterial(ctx, in)

		assert.Error(t, err)
		sts := status.Convert(err)
		assert.Equal(t, codes.Internal, sts.Code())
		assert.Contains(t, sts.Message(), "failed to get owner uuid: database error")
	})
	t.Run("user_not_owner", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("DeleteMaterial")
		mockLogger.EXPECT().Error("failed to delete: user is not owner")

		mockRepo.EXPECT().GetMaterialOwnerUUID(ctx, materialUUID).Return("other-uuid", nil)

		in := &materials.DeleteMaterialIn{Uuid: materialUUID}
		_, err := s.DeleteMaterial(ctx, in)

		assert.Error(t, err)
		sts := status.Convert(err)
		assert.Equal(t, codes.PermissionDenied, sts.Code())
		assert.Equal(t, "failed to delete: user is not owner", sts.Message())
	})

	t.Run("delete_material_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("DeleteMaterial")

		dbErr := fmt.Errorf("delete failed")

		mockLogger.EXPECT().Error(fmt.Sprintf("failed to delete material: %v", dbErr))

		mockRepo.EXPECT().GetMaterialOwnerUUID(ctx, materialUUID).Return(userUUID, nil)

		mockRepo.EXPECT().DeleteMaterial(ctx, gomock.Any()).Return(int64(0), dbErr)

		in := &materials.DeleteMaterialIn{Uuid: materialUUID}
		_, err := s.DeleteMaterial(ctx, in)

		assert.Error(t, err)
		sts := status.Convert(err)
		assert.Equal(t, codes.Internal, sts.Code())
		assert.Contains(t, sts.Message(), "failed to delete material: delete failed")
	})

	t.Run("material_not_found", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("DeleteMaterial")
		mockLogger.EXPECT().Error("failed to delete: material already deleted or not found")

		mockRepo.EXPECT().GetMaterialOwnerUUID(ctx, materialUUID).Return(userUUID, nil)
		mockRepo.EXPECT().DeleteMaterial(ctx, materialUUID).Return(int64(0), nil)

		in := &materials.DeleteMaterialIn{Uuid: materialUUID}
		_, err := s.DeleteMaterial(ctx, in)

		assert.Error(t, err)
		sts := status.Convert(err)
		assert.Equal(t, codes.NotFound, sts.Code())
		assert.Contains(t, sts.Message(), "failed to delete: material already deleted or not found")
	})
}

func TestServer_PublishMaterial(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockDBRepo(ctrl)
	mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

	userUUID := uuid.New().String()
	materialUUID := uuid.New().String()

	content := "material content"
	createdAt := time.Now()
	editedAt := time.Now()
	publishedAt := time.Now()

	ctx := context.Background()
	ctx = context.WithValue(ctx, config.KeyLogger, mockLogger)
	ctx = context.WithValue(ctx, config.KeyUUID, userUUID)

	s := New(mockRepo)

	t.Run("success", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("PublishMaterial")

		in := &materials.PublishMaterialIn{
			Uuid: materialUUID,
		}

		publishedMaterial := &model.Material{
			UUID:            materialUUID,
			OwnerUUID:       userUUID,
			Title:           "Test Material",
			CoverImageURL:   "http://example.com/cover.jpg",
			Description:     "Test description",
			Content:         &content,
			ReadTimeMinutes: 10,
			Status:          "published",
			CreatedAt:       createdAt,
			EditedAt:        &editedAt,
			PublishedAt:     &publishedAt,
			DeletedAt:       nil,
			LikesCount:      0,
		}

		mockRepo.EXPECT().GetMaterialOwnerUUID(ctx, materialUUID).Return(userUUID, nil)
		mockRepo.EXPECT().MaterialExists(ctx, materialUUID).Return(true, nil)
		mockRepo.EXPECT().PublishMaterial(ctx, materialUUID).Return(publishedMaterial, nil)

		out, err := s.PublishMaterial(ctx, in)

		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.Equal(t, publishedMaterial.FromDTO(), out.Material)
	})

	t.Run("empty_material_uuid", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("PublishMaterial")
		mockLogger.EXPECT().Error("material uuid is required")

		in := &materials.PublishMaterialIn{Uuid: ""}
		out, err := s.PublishMaterial(ctx, in)

		assert.Nil(t, out)
		assert.Error(t, err)
		sts := status.Convert(err)
		assert.Equal(t, codes.InvalidArgument, sts.Code())
		assert.Equal(t, "material uuid is required", sts.Message())
	})

	t.Run("missing_user_uuid", func(t *testing.T) {
		noCtx := context.Background()
		noCtx = context.WithValue(noCtx, config.KeyLogger, mockLogger)

		mockLogger.EXPECT().AddFuncName("PublishMaterial")
		mockLogger.EXPECT().Error("user uuid is required")

		in := &materials.PublishMaterialIn{Uuid: materialUUID}
		out, err := s.PublishMaterial(noCtx, in)

		assert.Nil(t, out)
		assert.Error(t, err)
		sts := status.Convert(err)
		assert.Equal(t, codes.Unauthenticated, sts.Code())
		assert.Equal(t, "user uuid is required", sts.Message())
	})

	t.Run("user_not_owner", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("PublishMaterial")
		mockLogger.EXPECT().Error("failed to publish: user is not owner")
		mockRepo.EXPECT().GetMaterialOwnerUUID(ctx, materialUUID).Return("other-uuid", nil)

		in := &materials.PublishMaterialIn{Uuid: materialUUID}
		out, err := s.PublishMaterial(ctx, in)

		assert.Nil(t, out)
		assert.Error(t, err)
		sts := status.Convert(err)
		assert.Equal(t, codes.PermissionDenied, sts.Code())
		assert.Equal(t, "failed to publish: user is not owner", sts.Message())
	})

	t.Run("get_material_owner_uuid_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("PublishMaterial")
		dbErr := fmt.Errorf("database error")
		mockLogger.EXPECT().Error(fmt.Sprintf("failed to get owner uuid: %v", dbErr))
		mockRepo.EXPECT().GetMaterialOwnerUUID(ctx, materialUUID).Return("", dbErr)

		in := &materials.PublishMaterialIn{Uuid: materialUUID}
		out, err := s.PublishMaterial(ctx, in)

		assert.Nil(t, out)
		assert.Error(t, err)
		sts := status.Convert(err)
		assert.Equal(t, codes.Internal, sts.Code())
		assert.Contains(t, sts.Message(), "failed to get owner uuid: database error")
	})
	t.Run("material_does_not_exist", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("PublishMaterial")
		mockLogger.EXPECT().Error("material does not exist")

		mockRepo.EXPECT().GetMaterialOwnerUUID(ctx, materialUUID).Return(userUUID, nil)
		mockRepo.EXPECT().MaterialExists(ctx, materialUUID).Return(false, nil)

		in := &materials.PublishMaterialIn{Uuid: materialUUID}
		out, err := s.PublishMaterial(ctx, in)

		assert.Nil(t, out)
		assert.Error(t, err)
		sts := status.Convert(err)
		assert.Equal(t, codes.FailedPrecondition, sts.Code())
		assert.Equal(t, "material does not exist", sts.Message())
	})

	t.Run("publish_material_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("PublishMaterial")
		dbErr := fmt.Errorf("publish failed")
		mockLogger.EXPECT().Error(fmt.Sprintf("failed to publish material: %v", dbErr))

		mockRepo.EXPECT().GetMaterialOwnerUUID(ctx, materialUUID).Return(userUUID, nil)
		mockRepo.EXPECT().MaterialExists(ctx, materialUUID).Return(true, nil)
		mockRepo.EXPECT().PublishMaterial(ctx, materialUUID).Return(nil, dbErr)

		in := &materials.PublishMaterialIn{Uuid: materialUUID}
		out, err := s.PublishMaterial(ctx, in)

		assert.Nil(t, out)
		assert.Error(t, err)
		sts := status.Convert(err)
		assert.Equal(t, codes.Internal, sts.Code())
		assert.Contains(t, sts.Message(), "failed to publish material: publish failed")
	})
}

func TestServer_ArchivedMaterial(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockDBRepo(ctrl)
	mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

	userUUID := uuid.New().String()
	materialUUID := uuid.New().String()

	ctx := context.Background()
	ctx = context.WithValue(ctx, config.KeyLogger, mockLogger)
	ctx = context.WithValue(ctx, config.KeyUUID, userUUID)

	s := New(mockRepo)

	t.Run("success", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("ArchivedMaterial")
		mockRepo.EXPECT().GetMaterialOwnerUUID(ctx, materialUUID).Return(userUUID, nil)
		mockRepo.EXPECT().ArchivedMaterial(ctx, materialUUID).Return(int64(1), nil)

		in := &materials.ArchivedMaterialIn{Uuid: materialUUID}

		out, err := s.ArchivedMaterial(ctx, in)

		assert.NoError(t, err)
		assert.Nil(t, out)
	})

	t.Run("no_user_uuid", func(t *testing.T) {
		badCtx := context.Background()
		badCtx = context.WithValue(badCtx, config.KeyLogger, mockLogger)
		mockLogger.EXPECT().AddFuncName("ArchivedMaterial")
		mockLogger.EXPECT().Error("uuid is required")

		in := &materials.ArchivedMaterialIn{Uuid: materialUUID}
		_, err := s.ArchivedMaterial(badCtx, in)

		assert.Error(t, err)
		sts := status.Convert(err)
		assert.Equal(t, codes.Unauthenticated, sts.Code())
		assert.Equal(t, "uuid is required", sts.Message())
	})

	t.Run("get_owner_uuid_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("ArchivedMaterial")

		dbErr := fmt.Errorf("database error")
		mockLogger.EXPECT().Error(fmt.Sprintf("failed to get owner uuid: %v", dbErr))
		mockRepo.EXPECT().GetMaterialOwnerUUID(ctx, materialUUID).Return("", dbErr)

		in := &materials.ArchivedMaterialIn{Uuid: materialUUID}
		_, err := s.ArchivedMaterial(ctx, in)

		assert.Error(t, err)
		sts := status.Convert(err)
		assert.Equal(t, codes.Internal, sts.Code())
		assert.Contains(t, sts.Message(), "failed to get owner uuid: database error")
	})
	t.Run("user_not_owner", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("ArchivedMaterial")
		mockLogger.EXPECT().Error("failed to archived: user is not owner")

		mockRepo.EXPECT().GetMaterialOwnerUUID(ctx, materialUUID).Return("other-uuid", nil)

		in := &materials.ArchivedMaterialIn{Uuid: materialUUID}
		_, err := s.ArchivedMaterial(ctx, in)
		assert.Error(t, err)
		sts := status.Convert(err)
		assert.Equal(t, codes.PermissionDenied, sts.Code())
		assert.Equal(t, "failed to archived: user is not owner", sts.Message())
	})

	t.Run("Archived_material_error", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("ArchivedMaterial")

		dbErr := fmt.Errorf("archived failed")

		mockLogger.EXPECT().Error(fmt.Sprintf("failed to archived material: %v", dbErr))

		mockRepo.EXPECT().GetMaterialOwnerUUID(ctx, materialUUID).Return(userUUID, nil)

		mockRepo.EXPECT().ArchivedMaterial(ctx, gomock.Any()).Return(int64(0), dbErr)

		in := &materials.ArchivedMaterialIn{Uuid: materialUUID}
		_, err := s.ArchivedMaterial(ctx, in)

		assert.Error(t, err)
		sts := status.Convert(err)
		assert.Equal(t, codes.Internal, sts.Code())
		assert.Contains(t, sts.Message(), "failed to archived material: archived failed")
	})

	t.Run("material_not_found", func(t *testing.T) {
		mockLogger.EXPECT().AddFuncName("ArchivedMaterial")
		mockLogger.EXPECT().Error("failed to archived material: material already archived or not found")

		mockRepo.EXPECT().GetMaterialOwnerUUID(ctx, materialUUID).Return(userUUID, nil)
		mockRepo.EXPECT().ArchivedMaterial(ctx, materialUUID).Return(int64(0), nil)

		in := &materials.ArchivedMaterialIn{Uuid: materialUUID}
		_, err := s.ArchivedMaterial(ctx, in)

		assert.Error(t, err)
		sts := status.Convert(err)
		assert.Equal(t, codes.NotFound, sts.Code())
		assert.Contains(t, sts.Message(), "failed to archived material: material already archived or not found")
	})
}

func TestService_ToggleLike(t *testing.T) {
	t.Parallel()
	baseCtx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockDBRepo(ctrl)
	mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

	userUUID := uuid.New().String()
	materialUUID := uuid.New().String()

	s := New(mockRepo)

	t.Run("add_like_successfully", func(t *testing.T) {
		ctx := context.WithValue(baseCtx, config.KeyLogger, mockLogger)
		ctx = context.WithValue(ctx, config.KeyUUID, userUUID)
		ctx = tx.WithDBRepoContext(ctx, mockRepo)

		in := &materials.ToggleLikeIn{MaterialUuid: materialUUID}

		mockLogger.EXPECT().AddFuncName("ToggleLike")
		mockRepo.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, cb func(context.Context) error) error {
			return cb(ctx)
		})
		mockRepo.EXPECT().CheckLike(gomock.Any(), materialUUID, userUUID).Return(false, nil)
		mockRepo.EXPECT().AddLike(gomock.Any(), materialUUID, userUUID).Return(true, nil)
		mockRepo.EXPECT().GetLikesCount(gomock.Any(), materialUUID).Return(int32(5), nil)
		mockRepo.EXPECT().UpdateLikesCount(gomock.Any(), materialUUID, int32(5)).Return(int32(5), nil)

		out, err := s.ToggleLike(ctx, in)

		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, out.IsLiked)
		assert.Equal(t, int32(5), out.LikesCount)
	})

	t.Run("remove_like_successfully", func(t *testing.T) {
		ctx := context.WithValue(baseCtx, config.KeyLogger, mockLogger)
		ctx = context.WithValue(ctx, config.KeyUUID, userUUID)
		ctx = tx.WithDBRepoContext(ctx, mockRepo)

		in := &materials.ToggleLikeIn{MaterialUuid: materialUUID}

		mockLogger.EXPECT().AddFuncName("ToggleLike")
		mockRepo.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, cb func(context.Context) error) error {
			return cb(ctx)
		})
		mockRepo.EXPECT().CheckLike(gomock.Any(), materialUUID, userUUID).Return(true, nil)
		mockRepo.EXPECT().RemoveLike(gomock.Any(), materialUUID, userUUID).Return(true, nil)
		mockRepo.EXPECT().GetLikesCount(gomock.Any(), materialUUID).Return(int32(4), nil)
		mockRepo.EXPECT().UpdateLikesCount(gomock.Any(), materialUUID, int32(4)).Return(int32(4), nil)

		out, err := s.ToggleLike(ctx, in)

		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.False(t, out.IsLiked)
		assert.Equal(t, int32(4), out.LikesCount)
	})

	t.Run("fail_if_uuid_missing", func(t *testing.T) {
		ctx := context.WithValue(baseCtx, config.KeyLogger, mockLogger)
		ctx = tx.WithDBRepoContext(ctx, mockRepo)

		in := &materials.ToggleLikeIn{MaterialUuid: materialUUID}

		mockLogger.EXPECT().AddFuncName("ToggleLike")
		mockLogger.EXPECT().Error("uuid is required")

		out, err := s.ToggleLike(ctx, in)

		assert.Nil(t, out)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Contains(t, st.Message(), "uuid is required")
	})

	t.Run("fail_check_like_error", func(t *testing.T) {
		ctx := context.WithValue(baseCtx, config.KeyLogger, mockLogger)
		ctx = context.WithValue(ctx, config.KeyUUID, userUUID)
		ctx = tx.WithDBRepoContext(ctx, mockRepo)

		dbErr := fmt.Errorf("database error")

		mockLogger.EXPECT().AddFuncName("ToggleLike")
		mockLogger.EXPECT().Error(fmt.Sprintf("transaction failed: failed to check like: %v", dbErr))
		mockRepo.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, cb func(context.Context) error) error {
			return cb(ctx)
		})
		mockRepo.EXPECT().CheckLike(gomock.Any(), materialUUID, userUUID).Return(false, dbErr)

		in := &materials.ToggleLikeIn{MaterialUuid: materialUUID}
		out, err := s.ToggleLike(ctx, in)

		assert.Nil(t, out)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Contains(t, st.Message(), "failed to check like: database error")
	})

	t.Run("fail_add_like_error", func(t *testing.T) {
		ctx := context.WithValue(baseCtx, config.KeyLogger, mockLogger)
		ctx = context.WithValue(ctx, config.KeyUUID, userUUID)
		ctx = tx.WithDBRepoContext(ctx, mockRepo)

		dbErr := fmt.Errorf("database error")

		mockLogger.EXPECT().AddFuncName("ToggleLike")
		mockLogger.EXPECT().Error(fmt.Sprintf("transaction failed: failed to add like: %v", dbErr))
		mockRepo.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, cb func(context.Context) error) error {
			return cb(ctx)
		})
		mockRepo.EXPECT().CheckLike(gomock.Any(), materialUUID, userUUID).Return(false, nil)
		mockRepo.EXPECT().AddLike(gomock.Any(), materialUUID, userUUID).Return(false, dbErr)

		in := &materials.ToggleLikeIn{MaterialUuid: materialUUID}
		out, err := s.ToggleLike(ctx, in)

		assert.Nil(t, out)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Contains(t, st.Message(), "failed to add like: database error")
	})

	t.Run("fail_remove_like_error", func(t *testing.T) {
		ctx := context.WithValue(baseCtx, config.KeyLogger, mockLogger)
		ctx = context.WithValue(ctx, config.KeyUUID, userUUID)
		ctx = tx.WithDBRepoContext(ctx, mockRepo)

		dbErr := fmt.Errorf("database error")

		mockLogger.EXPECT().AddFuncName("ToggleLike")
		mockLogger.EXPECT().Error(fmt.Sprintf("transaction failed: failed to remove like: %v", dbErr))
		mockRepo.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, cb func(context.Context) error) error {
			return cb(ctx)
		})
		mockRepo.EXPECT().CheckLike(gomock.Any(), materialUUID, userUUID).Return(true, nil)
		mockRepo.EXPECT().RemoveLike(gomock.Any(), materialUUID, userUUID).Return(false, dbErr)

		in := &materials.ToggleLikeIn{MaterialUuid: materialUUID}
		out, err := s.ToggleLike(ctx, in)

		assert.Nil(t, out)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Contains(t, st.Message(), "failed to remove like: database error")
	})

	t.Run("fail_get_likes_count_error", func(t *testing.T) {
		ctx := context.WithValue(baseCtx, config.KeyLogger, mockLogger)
		ctx = context.WithValue(ctx, config.KeyUUID, userUUID)
		ctx = tx.WithDBRepoContext(ctx, mockRepo)

		dbErr := fmt.Errorf("database error")

		mockLogger.EXPECT().AddFuncName("ToggleLike")
		mockLogger.EXPECT().Error(fmt.Sprintf("transaction failed: failed to get likes count: %v", dbErr))
		mockRepo.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, cb func(context.Context) error) error {
			return cb(ctx)
		})
		mockRepo.EXPECT().CheckLike(gomock.Any(), materialUUID, userUUID).Return(false, nil)
		mockRepo.EXPECT().AddLike(gomock.Any(), materialUUID, userUUID).Return(true, nil)
		mockRepo.EXPECT().GetLikesCount(gomock.Any(), materialUUID).Return(int32(0), dbErr)

		in := &materials.ToggleLikeIn{MaterialUuid: materialUUID}
		out, err := s.ToggleLike(ctx, in)

		assert.Nil(t, out)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Contains(t, st.Message(), "failed to get likes count: database error")
	})

	t.Run("fail_update_likes_count_error", func(t *testing.T) {
		ctx := context.WithValue(baseCtx, config.KeyLogger, mockLogger)
		ctx = context.WithValue(ctx, config.KeyUUID, userUUID)
		ctx = tx.WithDBRepoContext(ctx, mockRepo)

		dbErr := fmt.Errorf("database error")

		mockLogger.EXPECT().AddFuncName("ToggleLike")
		mockLogger.EXPECT().Error(fmt.Sprintf("transaction failed: failed to update likes count: %v", dbErr))
		mockRepo.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, cb func(context.Context) error) error {
			return cb(ctx)
		})
		mockRepo.EXPECT().CheckLike(gomock.Any(), materialUUID, userUUID).Return(false, nil)
		mockRepo.EXPECT().AddLike(gomock.Any(), materialUUID, userUUID).Return(true, nil)
		mockRepo.EXPECT().GetLikesCount(gomock.Any(), materialUUID).Return(int32(5), nil)
		mockRepo.EXPECT().UpdateLikesCount(gomock.Any(), materialUUID, int32(5)).Return(int32(0), dbErr)

		in := &materials.ToggleLikeIn{MaterialUuid: materialUUID}
		out, err := s.ToggleLike(ctx, in)

		assert.Nil(t, out)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Contains(t, st.Message(), "failed to update likes count: database error")
	})

	t.Run("fail_with_tx_error", func(t *testing.T) {
		ctx := context.WithValue(baseCtx, config.KeyLogger, mockLogger)
		ctx = context.WithValue(ctx, config.KeyUUID, userUUID)
		ctx = tx.WithDBRepoContext(ctx, mockRepo)

		dbErr := fmt.Errorf("transaction error")

		mockLogger.EXPECT().AddFuncName("ToggleLike")
		mockLogger.EXPECT().Error(fmt.Sprintf("transaction failed: %v", dbErr))
		mockRepo.EXPECT().WithTx(gomock.Any(), gomock.Any()).Return(dbErr)

		in := &materials.ToggleLikeIn{MaterialUuid: materialUUID}
		out, err := s.ToggleLike(ctx, in)

		assert.Nil(t, out)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Contains(t, st.Message(), "transaction failed: transaction error")
	})
}
