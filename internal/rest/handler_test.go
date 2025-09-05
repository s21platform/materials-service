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

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	logger_lib "github.com/s21platform/logger-lib"
	"github.com/s21platform/materials-service/internal/config"
	api "github.com/s21platform/materials-service/internal/generated"
	"github.com/s21platform/materials-service/internal/model"
)

func TestHandler_SaveDraftMaterial(t *testing.T) {
	t.Parallel()

	userUUID := uuid.New().String()
	materialUUID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockLogger.EXPECT().AddFuncName("SaveDraftMaterial")

		mockRepo.EXPECT().SaveDraftMaterial(gomock.Any(), userUUID, gomock.Any()).DoAndReturn(func(ctx context.Context, ownerUUID string, material *model.SaveDraftMaterial) (string, error) {
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

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", userUUID)

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.SaveDraftMaterial(w, req, api.SaveDraftMaterialParams{XUserID: userUUID})

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

		handler := &Handler{
			repository: mockRepo,
		}

		mockLogger.EXPECT().AddFuncName("SaveDraftMaterial")
		mockLogger.EXPECT().Error(gomock.Any())

		req := httptest.NewRequest(http.MethodPost, "/api/materials/draft", strings.NewReader("invalid json"))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", userUUID)

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.SaveDraftMaterial(w, req, api.SaveDraftMaterialParams{XUserID: userUUID})

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

		handler := &Handler{
			repository: mockRepo,
		}

		mockLogger.EXPECT().AddFuncName("SaveDraftMaterial")
		mockLogger.EXPECT().Error("failed to get user UUID")

		requestBody := api.SaveDraftMaterialIn{
			Title:           "Test Title",
			Content:         stringPtr("Test Content"),
			Description:     stringPtr("Test Description"),
			CoverImageUrl:   stringPtr("http://example.com/cover.jpg"),
			ReadTimeMinutes: int32Ptr(5),
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/draft", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", userUUID)

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.SaveDraftMaterial(w, req, api.SaveDraftMaterialParams{XUserID: userUUID})

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Message, "user UUID is required")
	})

	t.Run("user_uuid_mismatch", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockLogger.EXPECT().AddFuncName("SaveDraftMaterial")
		mockLogger.EXPECT().Error("user UUID mismatch with X-User-ID header")

		requestBody := api.SaveDraftMaterialIn{
			Title:           "Test Title",
			Content:         stringPtr("Test Content"),
			Description:     stringPtr("Test Description"),
			CoverImageUrl:   stringPtr("http://example.com/cover.jpg"),
			ReadTimeMinutes: int32Ptr(5),
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/draft", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", uuid.New().String())

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.SaveDraftMaterial(w, req, api.SaveDraftMaterialParams{XUserID: uuid.New().String()})

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Message, "user UUID mismatch")
	})

	t.Run("repository_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockLogger.EXPECT().AddFuncName("SaveDraftMaterial")
		mockLogger.EXPECT().Error(gomock.Any())

		mockRepo.EXPECT().SaveDraftMaterial(gomock.Any(), userUUID, gomock.Any()).Return("", fmt.Errorf("database error"))

		requestBody := api.SaveDraftMaterialIn{
			Title:           "Test Title",
			Content:         stringPtr("Test Content"),
			Description:     stringPtr("Test Description"),
			CoverImageUrl:   stringPtr("http://example.com/cover.jpg"),
			ReadTimeMinutes: int32Ptr(5),
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/draft", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", userUUID)

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.SaveDraftMaterial(w, req, api.SaveDraftMaterialParams{XUserID: userUUID})

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Message, "failed to save draft material")
	})
}

func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}
