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

func TestHandler_PublishMaterial(t *testing.T) {
	t.Parallel()

	userUUID := uuid.New().String()
	materialUUID := uuid.New().String()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockRepo.EXPECT().GetMaterialOwnerUUID(gomock.Any(), materialUUID).Return(userUUID, nil)
		mockRepo.EXPECT().MaterialExists(gomock.Any(), materialUUID).Return(true, nil)
		mockRepo.EXPECT().PublishMaterial(gomock.Any(), materialUUID).Return(&model.Material{
			UUID:            materialUUID,
			OwnerUUID:       userUUID,
			Title:           "Test Title",
			Content:         stringPtr("Test Content"),
			Description:     "Test Description",
			CoverImageURL:   "http://example.com/cover.jpg",
			ReadTimeMinutes: 5,
			Status:          "published",
			CreatedAt:       now,
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

	t.Run("material_exists_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

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
		handler := &Handler{repository: mockRepo}

		mockRepo.EXPECT().
			WithTx(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, cb func(ctx context.Context) error) error {
				return cb(ctx)
			})

		mockRepo.EXPECT().CheckLike(gomock.Any(), materialUUID, userUUID).Return(false, nil)
		mockRepo.EXPECT().AddLike(gomock.Any(), materialUUID, userUUID).Return(nil)
		mockRepo.EXPECT().GetLikesCount(gomock.Any(), materialUUID).Return(int32(10), nil)
		mockRepo.EXPECT().UpdateLikesCount(gomock.Any(), materialUUID, int32(10)).Return(nil)

		requestBody := api.ToggleLikeIn{MaterialUuid: materialUUID}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/toggle-like", bytes.NewReader(bodyBytes))

		reqCtx := createTxContext(req.Context(), mockRepo)
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
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
		assert.Equal(t, int32(10), response.LikesCount)
	})

	t.Run("success_remove_like", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)
		handler := &Handler{repository: mockRepo}

		mockRepo.EXPECT().
			WithTx(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, cb func(ctx context.Context) error) error {
				return cb(ctx)
			})

		mockRepo.EXPECT().CheckLike(gomock.Any(), materialUUID, userUUID).Return(true, nil)
		mockRepo.EXPECT().RemoveLike(gomock.Any(), materialUUID, userUUID).Return(nil)
		mockRepo.EXPECT().GetLikesCount(gomock.Any(), materialUUID).Return(int32(9), nil)
		mockRepo.EXPECT().UpdateLikesCount(gomock.Any(), materialUUID, int32(9)).Return(nil)

		requestBody := api.ToggleLikeIn{MaterialUuid: materialUUID}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/toggle-like", bytes.NewReader(bodyBytes))

		reqCtx := createTxContext(req.Context(), mockRepo)
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
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
		assert.Equal(t, int32(9), response.LikesCount)
	})

	t.Run("invalid_json", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)
		handler := &Handler{repository: mockRepo}

		req := httptest.NewRequest(http.MethodPost, "/api/materials/toggle-like", strings.NewReader("invalid json"))

		reqCtx := createTxContext(req.Context(), mockRepo)
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
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
		handler := &Handler{repository: mockRepo}

		requestBody := api.ToggleLikeIn{MaterialUuid: ""}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/toggle-like", bytes.NewReader(bodyBytes))

		reqCtx := createTxContext(req.Context(), mockRepo)
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
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
		assert.Contains(t, errorResp.Message, "material uuid is required")
	})

	t.Run("missing_user_uuid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)
		handler := &Handler{repository: mockRepo}

		requestBody := api.ToggleLikeIn{MaterialUuid: materialUUID}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/toggle-like", bytes.NewReader(bodyBytes))

		reqCtx := createTxContext(req.Context(), mockRepo)
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
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
		assert.Contains(t, errorResp.Message, "user UUID is required")
	})

	t.Run("repository_error_on_check_like", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)
		handler := &Handler{repository: mockRepo}

		mockRepo.EXPECT().
			WithTx(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, cb func(ctx context.Context) error) error {
				return cb(ctx)
			})

		mockRepo.EXPECT().CheckLike(gomock.Any(), materialUUID, userUUID).Return(false, fmt.Errorf("db error"))

		requestBody := api.ToggleLikeIn{MaterialUuid: materialUUID}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/toggle-like", bytes.NewReader(bodyBytes))

		reqCtx := createTxContext(req.Context(), mockRepo)
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
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

//func int32Ptr(i int32) *int32 {
//	return &i
//}
