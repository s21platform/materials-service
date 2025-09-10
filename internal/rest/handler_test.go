package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	logger_lib "github.com/s21platform/logger-lib"
	"github.com/s21platform/materials-service/internal/config"
	api "github.com/s21platform/materials-service/internal/generated"
	"github.com/s21platform/materials-service/internal/model"
	"github.com/s21platform/materials-service/internal/pkg/tx"
)

func createTxContext(ctx context.Context, mockRepo *MockDBRepo) context.Context {
	return context.WithValue(ctx, tx.KeyTx, tx.Tx{DbRepo: mockRepo})
}

func TestHandler_SaveDraftMaterial(t *testing.T) {
	t.Parallel()

	userUUID := uuid.New().String()
	materialUUID := uuid.New().String()
	title := "Draft Title"
	content := stringPtr("Draft Content")
	description := stringPtr("Draft Description")
	coverImage := stringPtr("http://example.com/cover.jpg")
	readTime := int32(10)

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockLogger.EXPECT().AddFuncName("SaveDraftMaterial")

		mockRepo.EXPECT().
			SaveDraftMaterial(gomock.Any(), userUUID, &model.SaveDraftMaterial{
				Title:           &title,
				Content:         content,
				Description:     description,
				CoverImageURL:   coverImage,
				ReadTimeMinutes: &readTime,
			}).
			Return(materialUUID, nil)

		requestBody := api.SaveDraftMaterialIn{
			Title:           title,
			Content:         content,
			Description:     description,
			CoverImageUrl:   coverImage,
			ReadTimeMinutes: &readTime,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/save-draft", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.SaveDraftMaterial(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.SaveDraftMaterialOut
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, materialUUID, response.Uuid)
	})

	t.Run("invalid_json", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{repository: mockRepo}

		mockLogger.EXPECT().AddFuncName("SaveDraftMaterial")
		mockLogger.EXPECT().Error(gomock.Any())

		req := httptest.NewRequest(http.MethodPost, "/api/materials/save-draft", strings.NewReader("invalid json"))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.SaveDraftMaterial(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Message, "invalid request body")
	})

	t.Run("missing_user_uuid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{repository: mockRepo}

		mockLogger.EXPECT().AddFuncName("SaveDraftMaterial")
		mockLogger.EXPECT().Error("failed to get user UUID")

		requestBody := api.SaveDraftMaterialIn{
			Title: title,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/save-draft", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		// intentionally do NOT set KeyUUID
		req = req.WithContext(reqCtx)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.SaveDraftMaterial(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Message, "user UUID is required")
	})

	t.Run("repository_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{repository: mockRepo}

		mockLogger.EXPECT().AddFuncName("SaveDraftMaterial")
		mockLogger.EXPECT().Error(gomock.Any())

		mockRepo.EXPECT().
			SaveDraftMaterial(gomock.Any(), userUUID, gomock.Any()).
			Return("", fmt.Errorf("database error"))

		requestBody := api.SaveDraftMaterialIn{
			Title: title,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/save-draft", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.SaveDraftMaterial(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Message, "failed to save draft material")
	})
}

func TestHandler_PublishMaterial(t *testing.T) {
	t.Parallel()

	userUUID := uuid.New().String()
	materialUUID := uuid.New().String()
	now := time.Now()

	test_title := "Test Title"
	status := "published"
	test_description := "Test Description"
	test_cover_image := "http://example.com/cover.jpg"
	test_read_time_minutes := int32(5)

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockLogger.EXPECT().AddFuncName("PublishMaterial")

		mockRepo.EXPECT().GetMaterialOwnerUUID(gomock.Any(), materialUUID).Return(userUUID, nil)
		mockRepo.EXPECT().MaterialExists(gomock.Any(), materialUUID).Return(true, nil)
		mockRepo.EXPECT().PublishMaterial(gomock.Any(), materialUUID).Return(&model.Material{
			UUID:            materialUUID,
			OwnerUUID:       userUUID,
			Title:           test_title,
			Content:         stringPtr("Test Content"),
			Description:     &test_description,
			CoverImageURL:   &test_cover_image,
			ReadTimeMinutes: &test_read_time_minutes,
			Status:          &status,
			CreatedAt:       &now,
		}, nil)

		requestBody := api.PublishMaterialIn{
			Uuid: materialUUID,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/publish-material", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.PublishMaterial(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.PublishMaterialOut
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, materialUUID, response.Material.Uuid)
		assert.Equal(t, userUUID, *response.Material.OwnerUuid)
		assert.Equal(t, "Test Title", response.Material.Title)
		assert.Equal(t, "Test Content", *response.Material.Content)
		assert.Equal(t, "Test Description", *response.Material.Description)
		assert.Equal(t, "http://example.com/cover.jpg", *response.Material.CoverImageUrl)
		assert.Equal(t, int32(5), *response.Material.ReadTimeMinutes)
		assert.Equal(t, "published", *response.Material.Status)
	})

	t.Run("invalid_json", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockLogger.EXPECT().AddFuncName("PublishMaterial")
		mockLogger.EXPECT().Error(gomock.Any())

		req := httptest.NewRequest(http.MethodPost, "/api/materials/publish-material", strings.NewReader("invalid json"))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.PublishMaterial(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Message, "invalid request body")
	})

	t.Run("missing_material_uuid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockLogger.EXPECT().AddFuncName("PublishMaterial")
		mockLogger.EXPECT().Error("material uuid is required")

		requestBody := api.PublishMaterialIn{
			Uuid: "",
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/publish-material", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.PublishMaterial(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Message, "material uuid is required")
	})

	t.Run("missing_user_uuid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockLogger.EXPECT().AddFuncName("PublishMaterial")
		mockLogger.EXPECT().Error("failed to get user UUID")

		requestBody := api.PublishMaterialIn{
			Uuid: materialUUID,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/publish-material", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.PublishMaterial(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Message, "user UUID is required")
	})

	t.Run("material_owner_mismatch", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockLogger.EXPECT().AddFuncName("PublishMaterial")
		mockLogger.EXPECT().Error("failed to publish: user is not owner")

		mockRepo.EXPECT().GetMaterialOwnerUUID(gomock.Any(), materialUUID).Return(uuid.New().String(), nil)

		requestBody := api.PublishMaterialIn{
			Uuid: materialUUID,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/publish-material", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.PublishMaterial(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Message, "failed to publish: user is not owner")
	})

	t.Run("material_does_not_exist", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockLogger.EXPECT().AddFuncName("PublishMaterial")
		mockLogger.EXPECT().Error("material does not exist")

		mockRepo.EXPECT().GetMaterialOwnerUUID(gomock.Any(), materialUUID).Return(userUUID, nil)
		mockRepo.EXPECT().MaterialExists(gomock.Any(), materialUUID).Return(false, nil)

		requestBody := api.PublishMaterialIn{
			Uuid: materialUUID,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/publish-material", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.PublishMaterial(w, req)

		assert.Equal(t, http.StatusPreconditionFailed, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Message, "material does not exist")
	})

	t.Run("get_owner_uuid_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockLogger.EXPECT().AddFuncName("PublishMaterial")
		mockLogger.EXPECT().Error(gomock.Any())

		mockRepo.EXPECT().GetMaterialOwnerUUID(gomock.Any(), materialUUID).Return("", fmt.Errorf("database error"))

		requestBody := api.PublishMaterialIn{
			Uuid: materialUUID,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/publish-material", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.PublishMaterial(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Message, "failed to get owner uuid")
	})

	t.Run("material_exists_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockLogger.EXPECT().AddFuncName("PublishMaterial")
		mockLogger.EXPECT().Error(gomock.Any())

		mockRepo.EXPECT().GetMaterialOwnerUUID(gomock.Any(), materialUUID).Return(userUUID, nil)
		mockRepo.EXPECT().MaterialExists(gomock.Any(), materialUUID).Return(false, fmt.Errorf("database error"))

		requestBody := api.PublishMaterialIn{
			Uuid: materialUUID,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/publish-material", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.PublishMaterial(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Message, "failed to check material existence")
	})

	t.Run("publish_material_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockLogger.EXPECT().AddFuncName("PublishMaterial")
		mockLogger.EXPECT().Error(gomock.Any())

		mockRepo.EXPECT().GetMaterialOwnerUUID(gomock.Any(), materialUUID).Return(userUUID, nil)
		mockRepo.EXPECT().MaterialExists(gomock.Any(), materialUUID).Return(true, nil)
		mockRepo.EXPECT().PublishMaterial(gomock.Any(), materialUUID).Return(nil, fmt.Errorf("database error"))

		requestBody := api.PublishMaterialIn{
			Uuid: materialUUID,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/publish-material", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.PublishMaterial(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Message, "failed to publish material")
	})
}

func TestHandler_ToggleLike(t *testing.T) {
	t.Parallel()

	userUUID := uuid.New().String()
	materialUUID := uuid.New().String()

	t.Run("success_add_like", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockLogger.EXPECT().AddFuncName("ToggleLike")

		mockRepo.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, cb func(ctx context.Context) error) error {
			return cb(ctx)
		})

		mockRepo.EXPECT().CheckLike(gomock.Any(), materialUUID, userUUID).Return(false, nil)
		mockRepo.EXPECT().AddLike(gomock.Any(), materialUUID, userUUID).Return(nil)
		mockRepo.EXPECT().GetLikesCount(gomock.Any(), materialUUID).Return(int32(1), nil)
		mockRepo.EXPECT().UpdateLikesCount(gomock.Any(), materialUUID, int32(1)).Return(nil)

		requestBody := api.ToggleLikeIn{
			MaterialUuid: materialUUID,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPut, "/api/materials/toggle-like", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		reqCtx = createTxContext(reqCtx, mockRepo)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ToggleLike(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.ToggleLikeOut
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.IsLiked)
		assert.Equal(t, int32(1), response.LikesCount)
	})

	t.Run("success_remove_like", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockLogger.EXPECT().AddFuncName("ToggleLike")

		mockRepo.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, cb func(ctx context.Context) error) error {
			return cb(ctx)
		})

		mockRepo.EXPECT().CheckLike(gomock.Any(), materialUUID, userUUID).Return(true, nil)
		mockRepo.EXPECT().RemoveLike(gomock.Any(), materialUUID, userUUID).Return(nil)
		mockRepo.EXPECT().GetLikesCount(gomock.Any(), materialUUID).Return(int32(0), nil)
		mockRepo.EXPECT().UpdateLikesCount(gomock.Any(), materialUUID, int32(0)).Return(nil)

		requestBody := api.ToggleLikeIn{
			MaterialUuid: materialUUID,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPut, "/api/materials/toggle-like", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		reqCtx = createTxContext(reqCtx, mockRepo)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ToggleLike(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.ToggleLikeOut
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.IsLiked)
		assert.Equal(t, int32(0), response.LikesCount)
	})

	t.Run("invalid_json", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockLogger.EXPECT().AddFuncName("ToggleLike")
		mockLogger.EXPECT().Error(gomock.Any())

		req := httptest.NewRequest(http.MethodPut, "/api/materials/toggle-like", strings.NewReader("invalid json"))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		reqCtx = createTxContext(reqCtx, mockRepo)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ToggleLike(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Message, "invalid request body")
	})

	t.Run("missing_material_uuid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockLogger.EXPECT().AddFuncName("ToggleLike")
		mockLogger.EXPECT().Error("material uuid is required")

		requestBody := api.ToggleLikeIn{
			MaterialUuid: "",
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPut, "/api/materials/toggle-like", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		reqCtx = createTxContext(reqCtx, mockRepo)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ToggleLike(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Equal(t, "material uuid is required", errorResp.Message)
	})

	t.Run("missing_user_uuid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockLogger.EXPECT().AddFuncName("ToggleLike")
		mockLogger.EXPECT().Error("failed to get user UUID")

		requestBody := api.ToggleLikeIn{
			MaterialUuid: materialUUID,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPut, "/api/materials/toggle-like", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = createTxContext(reqCtx, mockRepo)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ToggleLike(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Equal(t, "user UUID is required", errorResp.Message)
	})

	t.Run("toggle_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockLogger.EXPECT().AddFuncName("ToggleLike")
		mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

		mockRepo.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, cb func(ctx context.Context) error) error {
			return cb(ctx)
		})

		mockRepo.EXPECT().CheckLike(gomock.Any(), materialUUID, userUUID).Return(false, fmt.Errorf("database error"))

		requestBody := api.ToggleLikeIn{
			MaterialUuid: materialUUID,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPut, "/api/materials/toggle-like", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		reqCtx = createTxContext(reqCtx, mockRepo)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ToggleLike(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Message, "failed to toggle like")
	})

	t.Run("transaction_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockLogger.EXPECT().AddFuncName("ToggleLike")
		mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

		mockRepo.EXPECT().WithTx(gomock.Any(), gomock.Any()).Return(fmt.Errorf("transaction failed"))

		requestBody := api.ToggleLikeIn{
			MaterialUuid: materialUUID,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPut, "/api/materials/toggle-like", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		reqCtx = createTxContext(reqCtx, mockRepo)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ToggleLike(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Message, "failed to toggle like")
	})
}

func stringPtr(s string) *string {
	return &s
}
