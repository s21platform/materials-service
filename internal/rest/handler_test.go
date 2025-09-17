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
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/s21platform/materials-service/internal/config"
	api "github.com/s21platform/materials-service/internal/generated"
	"github.com/s21platform/materials-service/internal/model"
	"github.com/s21platform/materials-service/internal/pkg/tx"
)

func createTxContext(ctx context.Context, mockRepo *MockDBRepo) context.Context {
	// Use a real *sqlx.Tx or a mock if needed; here we use a placeholder
	mockTx := &sqlx.Tx{}
	return context.WithValue(ctx, tx.KeyTx, mockTx)
}

func TestHandler_SaveDraftMaterial(t *testing.T) {
	t.Parallel()

	userUUID := uuid.New().String()
	materialUUID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := NewMockDBRepo(ctrl)
		handler := &Handler{repository: mockRepo}

		mockRepo.EXPECT().SaveDraftMaterial(gomock.Any(), userUUID, gomock.Any()).
			DoAndReturn(func(ctx context.Context, ownerUUID string, material *model.SaveDraftMaterial) (string, error) {
				assert.Equal(t, "Test Title", material.Title)
				assert.Equal(t, "Test Content", material.Content)
				assert.Equal(t, "Test Description", material.Description)
				assert.Equal(t, "http://example.com/cover.jpg", material.CoverImageURL)
				assert.Equal(t, int32(5), material.ReadTimeMinutes)
				return materialUUID, nil
			})

		requestBody := api.SaveDraftMaterialIn{
			Title:           "Test Title",
			Content:         stringPtr("Test Content"),
			Description:     stringPtr("Test Description"),
			CoverImageUrl:   stringPtr("http://example.com/cover.jpg"),
			ReadTimeMinutes: int32Ptr(5),
		}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/draft", bytes.NewReader(bodyBytes))
		reqCtx := context.WithValue(req.Context(), config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chi.NewRouteContext()))

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
		handler := &Handler{repository: mockRepo}

		req := httptest.NewRequest(http.MethodPost, "/api/materials/draft", strings.NewReader("invalid json"))
		reqCtx := context.WithValue(req.Context(), config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		w := httptest.NewRecorder()
		handler.SaveDraftMaterial(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing_user_uuid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := NewMockDBRepo(ctrl)
		handler := &Handler{repository: mockRepo}

		requestBody := api.SaveDraftMaterialIn{
			Title:           "Test Title",
			Content:         stringPtr("Test Content"),
			Description:     stringPtr("Test Description"),
			CoverImageUrl:   stringPtr("http://example.com/cover.jpg"),
			ReadTimeMinutes: int32Ptr(5),
		}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/draft", bytes.NewReader(bodyBytes))

		w := httptest.NewRecorder()
		handler.SaveDraftMaterial(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("repository_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := NewMockDBRepo(ctrl)
		handler := &Handler{repository: mockRepo}

		mockRepo.EXPECT().SaveDraftMaterial(gomock.Any(), userUUID, gomock.Any()).
			Return("", fmt.Errorf("database error"))

		requestBody := api.SaveDraftMaterialIn{
			Title:           "Test Title",
			Content:         stringPtr("Test Content"),
			Description:     stringPtr("Test Description"),
			CoverImageUrl:   stringPtr("http://example.com/cover.jpg"),
			ReadTimeMinutes: int32Ptr(5),
		}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/draft", bytes.NewReader(bodyBytes))
		reqCtx := context.WithValue(req.Context(), config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		w := httptest.NewRecorder()
		handler.SaveDraftMaterial(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
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
		handler := &Handler{repository: mockRepo}

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

		requestBody := api.PublishMaterialIn{Uuid: materialUUID}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/publish-material", bytes.NewReader(bodyBytes))
		reqCtx := context.WithValue(req.Context(), config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		w := httptest.NewRecorder()
		handler.PublishMaterial(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.PublishMaterialOut
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, materialUUID, response.Material.Uuid)
		assert.Equal(t, userUUID, *response.Material.OwnerUuid)
	})

	t.Run("invalid_json", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := NewMockDBRepo(ctrl)
		handler := &Handler{repository: mockRepo}

		req := httptest.NewRequest(http.MethodPost, "/api/materials/publish-material", strings.NewReader("invalid json"))
		reqCtx := context.WithValue(req.Context(), config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		w := httptest.NewRecorder()
		handler.PublishMaterial(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing_material_uuid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := NewMockDBRepo(ctrl)
		handler := &Handler{repository: mockRepo}

		requestBody := api.PublishMaterialIn{Uuid: ""}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/publish-material", bytes.NewReader(bodyBytes))
		reqCtx := context.WithValue(req.Context(), config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		w := httptest.NewRecorder()
		handler.PublishMaterial(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing_user_uuid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := NewMockDBRepo(ctrl)
		handler := &Handler{repository: mockRepo}

		requestBody := api.PublishMaterialIn{Uuid: materialUUID}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/publish-material", bytes.NewReader(bodyBytes))

		w := httptest.NewRecorder()
		handler.PublishMaterial(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("material_owner_mismatch", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := NewMockDBRepo(ctrl)
		handler := &Handler{repository: mockRepo}

		mockRepo.EXPECT().GetMaterialOwnerUUID(gomock.Any(), materialUUID).Return(uuid.New().String(), nil)

		requestBody := api.PublishMaterialIn{Uuid: materialUUID}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/publish-material", bytes.NewReader(bodyBytes))
		reqCtx := context.WithValue(req.Context(), config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		w := httptest.NewRecorder()
		handler.PublishMaterial(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("material_does_not_exist", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := NewMockDBRepo(ctrl)
		handler := &Handler{repository: mockRepo}

		mockRepo.EXPECT().GetMaterialOwnerUUID(gomock.Any(), materialUUID).Return(userUUID, nil)
		mockRepo.EXPECT().MaterialExists(gomock.Any(), materialUUID).Return(false, nil)

		requestBody := api.PublishMaterialIn{Uuid: materialUUID}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/publish-material", bytes.NewReader(bodyBytes))
		reqCtx := context.WithValue(req.Context(), config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		w := httptest.NewRecorder()
		handler.PublishMaterial(w, req)

		assert.Equal(t, http.StatusPreconditionFailed, w.Code)
	})

	t.Run("get_owner_uuid_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := NewMockDBRepo(ctrl)
		handler := &Handler{repository: mockRepo}

		mockRepo.EXPECT().GetMaterialOwnerUUID(gomock.Any(), materialUUID).Return("", fmt.Errorf("database error"))

		requestBody := api.PublishMaterialIn{Uuid: materialUUID}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/publish-material", bytes.NewReader(bodyBytes))
		reqCtx := context.WithValue(req.Context(), config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		w := httptest.NewRecorder()
		handler.PublishMaterial(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("material_exists_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := NewMockDBRepo(ctrl)
		handler := &Handler{repository: mockRepo}

		mockRepo.EXPECT().GetMaterialOwnerUUID(gomock.Any(), materialUUID).Return(userUUID, nil)
		mockRepo.EXPECT().MaterialExists(gomock.Any(), materialUUID).Return(false, fmt.Errorf("database error"))

		requestBody := api.PublishMaterialIn{Uuid: materialUUID}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/publish-material", bytes.NewReader(bodyBytes))
		reqCtx := context.WithValue(req.Context(), config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		w := httptest.NewRecorder()
		handler.PublishMaterial(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("publish_material_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := NewMockDBRepo(ctrl)
		handler := &Handler{repository: mockRepo}

		mockRepo.EXPECT().GetMaterialOwnerUUID(gomock.Any(), materialUUID).Return(userUUID, nil)
		mockRepo.EXPECT().MaterialExists(gomock.Any(), materialUUID).Return(true, nil)
		mockRepo.EXPECT().PublishMaterial(gomock.Any(), materialUUID).Return(nil, fmt.Errorf("database error"))

		requestBody := api.PublishMaterialIn{Uuid: materialUUID}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/publish-material", bytes.NewReader(bodyBytes))
		reqCtx := context.WithValue(req.Context(), config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		w := httptest.NewRecorder()
		handler.PublishMaterial(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
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
		handler := &Handler{repository: mockRepo}

		// Mock WithTx to simulate transaction execution
		mockRepo.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, cb func(context.Context) error) error {
				// Simulate setting *sqlx.Tx in the context
				newCtx := context.WithValue(ctx, tx.KeyTx, &sqlx.Tx{})
				return cb(newCtx)
			},
		).Return(nil)

		mockRepo.EXPECT().CheckLike(gomock.Any(), materialUUID, userUUID).Return(false, nil)
		mockRepo.EXPECT().AddLike(gomock.Any(), materialUUID, userUUID).Return(nil)
		mockRepo.EXPECT().GetLikesCount(gomock.Any(), materialUUID).Return(int32(10), nil)
		mockRepo.EXPECT().UpdateLikesCount(gomock.Any(), materialUUID, int32(10)).Return(nil)

		requestBody := api.ToggleLikeIn{MaterialUuid: materialUUID}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/toggle-like", bytes.NewReader(bodyBytes))

		reqCtx := createTxContext(req.Context(), mockRepo)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

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
		handler := &Handler{repository: mockRepo}

		mockRepo.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, cb func(context.Context) error) error {
				newCtx := context.WithValue(ctx, tx.KeyTx, &sqlx.Tx{})
				return cb(newCtx)
			},
		).Return(nil)

		mockRepo.EXPECT().CheckLike(gomock.Any(), materialUUID, userUUID).Return(true, nil)
		mockRepo.EXPECT().RemoveLike(gomock.Any(), materialUUID, userUUID).Return(nil)
		mockRepo.EXPECT().GetLikesCount(gomock.Any(), materialUUID).Return(int32(9), nil)
		mockRepo.EXPECT().UpdateLikesCount(gomock.Any(), materialUUID, int32(9)).Return(nil)

		requestBody := api.ToggleLikeIn{MaterialUuid: materialUUID}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/toggle-like", bytes.NewReader(bodyBytes))

		reqCtx := createTxContext(req.Context(), mockRepo)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

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
		handler := &Handler{repository: mockRepo}

		req := httptest.NewRequest(http.MethodPost, "/api/materials/toggle-like", strings.NewReader("invalid json"))
		reqCtx := createTxContext(req.Context(), mockRepo)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		w := httptest.NewRecorder()
		handler.ToggleLike(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing_material_uuid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := NewMockDBRepo(ctrl)
		handler := &Handler{repository: mockRepo}

		requestBody := api.ToggleLikeIn{MaterialUuid: ""}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/toggle-like", bytes.NewReader(bodyBytes))
		reqCtx := createTxContext(req.Context(), mockRepo)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		w := httptest.NewRecorder()
		handler.ToggleLike(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing_user_uuid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := NewMockDBRepo(ctrl)
		handler := &Handler{repository: mockRepo}

		requestBody := api.ToggleLikeIn{MaterialUuid: materialUUID}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/toggle-like", bytes.NewReader(bodyBytes))
		reqCtx := createTxContext(req.Context(), mockRepo)
		req = req.WithContext(reqCtx)

		w := httptest.NewRecorder()
		handler.ToggleLike(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("repository_error_on_check_like", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := NewMockDBRepo(ctrl)
		handler := &Handler{repository: mockRepo}

		mockRepo.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, cb func(context.Context) error) error {
				newCtx := context.WithValue(ctx, tx.KeyTx, &sqlx.Tx{})
				return cb(newCtx)
			},
		).Return(nil)

		mockRepo.EXPECT().CheckLike(gomock.Any(), materialUUID, userUUID).Return(false, fmt.Errorf("db error"))

		requestBody := api.ToggleLikeIn{MaterialUuid: materialUUID}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/toggle-like", bytes.NewReader(bodyBytes))
		reqCtx := createTxContext(req.Context(), mockRepo)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		w := httptest.NewRecorder()
		handler.ToggleLike(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func stringPtr(s string) *string { return &s }
func int32Ptr(i int32) *int32    { return &i }
