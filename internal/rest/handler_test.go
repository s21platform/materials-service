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
	"google.golang.org/protobuf/types/known/timestamppb"

	logger_lib "github.com/s21platform/logger-lib"
	"github.com/s21platform/materials-service/internal/config"
	api "github.com/s21platform/materials-service/internal/generated"
	"github.com/s21platform/materials-service/internal/model"
	"github.com/s21platform/materials-service/internal/pkg/tx"
	proto "github.com/s21platform/materials-service/pkg/materials"
)

func createTxContext(ctx context.Context, mockRepo *MockDBRepo) context.Context {
	return context.WithValue(ctx, tx.KeyTx, tx.Tx{DbRepo: mockRepo})
}

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

		title := "Test Title"
		content := "Test Content"
		description := "Test Description"
		coverImageURL := "http://example.com/cover.jpg"
		readTimeMinutes := int32(5)

		expectedMaterial := &model.SaveDraftMaterial{
			Title:           title,
			Content:         content,
			Description:     description,
			CoverImageURL:   coverImageURL,
			ReadTimeMinutes: readTimeMinutes,
		}

		mockRepo.EXPECT().SaveDraftMaterial(gomock.Any(), userUUID, expectedMaterial).Return(materialUUID, nil)

		requestBody := api.SaveDraftMaterialIn{
			Title:           title,
			Content:         content,
			Description:     description,
			CoverImageUrl:   coverImageURL,
			ReadTimeMinutes: readTimeMinutes,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/save-draft-material", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

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

		handler := &Handler{
			repository: mockRepo,
		}

		req := httptest.NewRequest(http.MethodPost, "/api/materials/save-draft-material", strings.NewReader("invalid json"))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

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

		handler := &Handler{
			repository: mockRepo,
		}

		title := "Test Title"
		content := "Test Content"
		description := "Test Description"
		coverImageURL := "http://example.com/cover.jpg"
		readTimeMinutes := int32(5)

		requestBody := api.SaveDraftMaterialIn{
			Title:           title,
			Content:         content,
			Description:     description,
			CoverImageUrl:   coverImageURL,
			ReadTimeMinutes: readTimeMinutes,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/save-draft-material", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

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

		handler := &Handler{
			repository: mockRepo,
		}

		title := "Test Title"
		content := "Test Content"
		description := "Test Description"
		coverImageURL := "http://example.com/cover.jpg"
		readTimeMinutes := int32(5)

		expectedMaterial := &model.SaveDraftMaterial{
			Title:           title,
			Content:         content,
			Description:     description,
			CoverImageURL:   coverImageURL,
			ReadTimeMinutes: readTimeMinutes,
		}

		mockRepo.EXPECT().SaveDraftMaterial(gomock.Any(), userUUID, expectedMaterial).Return("", fmt.Errorf("db error"))

		requestBody := api.SaveDraftMaterialIn{
			Title:           title,
			Content:         content,
			Description:     description,
			CoverImageUrl:   coverImageURL,
			ReadTimeMinutes: readTimeMinutes,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/save-draft-material", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

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

	description := "Test Description"
	coverImageUrl := "http://example.com/cover.jpg"
	readTimeMinutes := int32(5)

	t.Run("success with kafka produce", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)
		mockKafka := NewMockKafkaProducer(ctrl)

		handler := &Handler{
			repository:          mockRepo,
			createKafkaProducer: mockKafka,
		}

		mockRepo.EXPECT().GetMaterialOwnerUUID(gomock.Any(), materialUUID).Return(userUUID, nil)
		mockRepo.EXPECT().MaterialExists(gomock.Any(), materialUUID).Return(true, nil)
		mockRepo.EXPECT().PublishMaterial(gomock.Any(), materialUUID).Return(&model.Material{
			UUID:            materialUUID,
			OwnerUUID:       userUUID,
			Title:           "Test Title",
			Content:         stringPtr("Test Content"),
			Description:     description,
			CoverImageURL:   coverImageUrl,
			ReadTimeMinutes: readTimeMinutes,
			Status:          "published",
			CreatedAt:       now,
		}, nil)

		mockKafka.EXPECT().
			ProduceMessage(gomock.Any(), gomock.Any(), userUUID).
			Return(nil)

		requestBody := api.PublishMaterialIn{Uuid: materialUUID}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/publish-material", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		reqCtx := context.WithValue(req.Context(), config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

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
		assert.Equal(t, "Test Content", response.Material.Content)
		assert.Equal(t, "Test Description", response.Material.Description)
		assert.Equal(t, "http://example.com/cover.jpg", response.Material.CoverImageUrl)
		assert.Equal(t, int32(5), response.Material.ReadTimeMinutes)
		assert.Equal(t, "published", response.Material.Status)
	})

	t.Run("success but kafka produce fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)
		mockKafka := NewMockKafkaProducer(ctrl)

		handler := &Handler{
			repository:          mockRepo,
			createKafkaProducer: mockKafka,
		}

		mockRepo.EXPECT().GetMaterialOwnerUUID(gomock.Any(), materialUUID).Return(userUUID, nil)
		mockRepo.EXPECT().MaterialExists(gomock.Any(), materialUUID).Return(true, nil)
		mockRepo.EXPECT().PublishMaterial(gomock.Any(), materialUUID).Return(&model.Material{
			UUID:            materialUUID,
			OwnerUUID:       userUUID,
			Title:           "Test Title",
			Content:         stringPtr("Test Content"),
			Description:     description,
			CoverImageURL:   coverImageUrl,
			ReadTimeMinutes: readTimeMinutes,
			Status:          "published",
			CreatedAt:       now,
		}, nil)

		mockKafka.EXPECT().
			ProduceMessage(gomock.Any(), gomock.Any(), userUUID).
			Return(fmt.Errorf("kafka error"))

		requestBody := api.PublishMaterialIn{Uuid: materialUUID}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/publish-material", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		reqCtx := context.WithValue(req.Context(), config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.PublishMaterial(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid_json", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)
		mockKafka := NewMockKafkaProducer(ctrl)

		handler := &Handler{
			repository:          mockRepo,
			createKafkaProducer: mockKafka,
		}

		req := httptest.NewRequest(http.MethodPost, "/api/materials/publish-material", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		reqCtx := context.WithValue(req.Context(), config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.PublishMaterial(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing_material_uuid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)
		mockKafka := NewMockKafkaProducer(ctrl)

		handler := &Handler{
			repository:          mockRepo,
			createKafkaProducer: mockKafka,
		}

		requestBody := api.PublishMaterialIn{Uuid: ""}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/publish-material", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		reqCtx := context.WithValue(req.Context(), config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.PublishMaterial(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing_user_uuid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)
		mockKafka := NewMockKafkaProducer(ctrl)

		handler := &Handler{
			repository:          mockRepo,
			createKafkaProducer: mockKafka,
		}

		requestBody := api.PublishMaterialIn{Uuid: materialUUID}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/publish-material", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		reqCtx := context.WithValue(req.Context(), config.KeyLogger, mockLogger)

		req = req.WithContext(reqCtx)

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.PublishMaterial(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("material_owner_mismatch", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

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

	t.Run("material_exists_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)
		mockKafka := NewMockKafkaProducer(ctrl)

		handler := &Handler{
			repository:          mockRepo,
			createKafkaProducer: mockKafka,
		}

		mockRepo.EXPECT().GetMaterialOwnerUUID(gomock.Any(), materialUUID).Return(userUUID, nil)
		mockRepo.EXPECT().MaterialExists(gomock.Any(), materialUUID).Return(false, fmt.Errorf("database error"))

		requestBody := api.PublishMaterialIn{Uuid: materialUUID}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/publish-material", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		reqCtx := context.WithValue(req.Context(), config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.PublishMaterial(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("publish_material_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)
		mockKafka := NewMockKafkaProducer(ctrl)

		handler := &Handler{
			repository:          mockRepo,
			createKafkaProducer: mockKafka,
		}

		mockRepo.EXPECT().GetMaterialOwnerUUID(gomock.Any(), materialUUID).Return(userUUID, nil)
		mockRepo.EXPECT().MaterialExists(gomock.Any(), materialUUID).Return(true, nil)
		mockRepo.EXPECT().PublishMaterial(gomock.Any(), materialUUID).Return(nil, fmt.Errorf("database error"))

		requestBody := api.PublishMaterialIn{Uuid: materialUUID}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/publish-material", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		reqCtx := context.WithValue(req.Context(), config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

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
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)
		mockLikeKafka := NewMockKafkaProducer(ctrl)

		handler := &Handler{
			repository:        mockRepo,
			likeKafkaProducer: mockLikeKafka,
		}

		mockRepo.EXPECT().
			WithTx(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, cb func(ctx context.Context) error) error {
				return cb(ctx)
			})

		mockRepo.EXPECT().CheckLike(gomock.Any(), materialUUID, userUUID).Return(false, nil)
		mockRepo.EXPECT().AddLike(gomock.Any(), materialUUID, userUUID).Return(nil)
		mockRepo.EXPECT().GetLikesCount(gomock.Any(), materialUUID).Return(int32(10), nil)
		mockRepo.EXPECT().UpdateLikesCount(gomock.Any(), materialUUID, int32(10)).Return(nil)

		likeMsg := &proto.ToggleLikeMessage{
			MaterialUuid: materialUUID,
			IsLiked:      true,
			LikesCount:   10,
		}
		mockLikeKafka.EXPECT().
			ProduceMessage(gomock.Any(), likeMsg, materialUUID).
			Return(nil)

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

	t.Run("success_add_like_with_kafka_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)
		mockLikeKafka := NewMockKafkaProducer(ctrl)

		handler := &Handler{
			repository:        mockRepo,
			likeKafkaProducer: mockLikeKafka,
		}

		mockRepo.EXPECT().
			WithTx(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, cb func(ctx context.Context) error) error {
				return cb(ctx)
			})

		mockRepo.EXPECT().CheckLike(gomock.Any(), materialUUID, userUUID).Return(false, nil)
		mockRepo.EXPECT().AddLike(gomock.Any(), materialUUID, userUUID).Return(nil)
		mockRepo.EXPECT().GetLikesCount(gomock.Any(), materialUUID).Return(int32(10), nil)
		mockRepo.EXPECT().UpdateLikesCount(gomock.Any(), materialUUID, int32(10)).Return(nil)

		likeMsg := &proto.ToggleLikeMessage{
			MaterialUuid: materialUUID,
			IsLiked:      true,
			LikesCount:   10,
		}
		mockLikeKafka.EXPECT().
			ProduceMessage(gomock.Any(), likeMsg, materialUUID).
			Return(fmt.Errorf("kafka error"))

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
		mockLikeKafka := NewMockKafkaProducer(ctrl)

		handler := &Handler{
			repository:        mockRepo,
			likeKafkaProducer: mockLikeKafka,
		}

		mockRepo.EXPECT().
			WithTx(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, cb func(ctx context.Context) error) error {
				return cb(ctx)
			})

		mockRepo.EXPECT().CheckLike(gomock.Any(), materialUUID, userUUID).Return(true, nil)
		mockRepo.EXPECT().RemoveLike(gomock.Any(), materialUUID, userUUID).Return(nil)
		mockRepo.EXPECT().GetLikesCount(gomock.Any(), materialUUID).Return(int32(9), nil)
		mockRepo.EXPECT().UpdateLikesCount(gomock.Any(), materialUUID, int32(9)).Return(nil)

		likeMsg := &proto.ToggleLikeMessage{
			MaterialUuid: materialUUID,
			IsLiked:      false,
			LikesCount:   9,
		}
		mockLikeKafka.EXPECT().
			ProduceMessage(gomock.Any(), likeMsg, materialUUID).
			Return(nil)

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

	t.Run("success_remove_like_with_kafka_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)
		mockLikeKafka := NewMockKafkaProducer(ctrl)

		handler := &Handler{
			repository:        mockRepo,
			likeKafkaProducer: mockLikeKafka,
		}

		mockRepo.EXPECT().
			WithTx(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, cb func(ctx context.Context) error) error {
				return cb(ctx)
			})

		mockRepo.EXPECT().CheckLike(gomock.Any(), materialUUID, userUUID).Return(true, nil)
		mockRepo.EXPECT().RemoveLike(gomock.Any(), materialUUID, userUUID).Return(nil)
		mockRepo.EXPECT().GetLikesCount(gomock.Any(), materialUUID).Return(int32(9), nil)
		mockRepo.EXPECT().UpdateLikesCount(gomock.Any(), materialUUID, int32(9)).Return(nil)

		likeMsg := &proto.ToggleLikeMessage{
			MaterialUuid: materialUUID,
			IsLiked:      false,
			LikesCount:   9,
		}
		mockLikeKafka.EXPECT().
			ProduceMessage(gomock.Any(), likeMsg, materialUUID).
			Return(fmt.Errorf("kafka error"))

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
		mockLikeKafka := NewMockKafkaProducer(ctrl)

		handler := &Handler{
			repository:        mockRepo,
			likeKafkaProducer: mockLikeKafka,
		}

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
		mockLikeKafka := NewMockKafkaProducer(ctrl)

		handler := &Handler{
			repository:        mockRepo,
			likeKafkaProducer: mockLikeKafka,
		}

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
		mockLikeKafka := NewMockKafkaProducer(ctrl)

		handler := &Handler{
			repository:        mockRepo,
			likeKafkaProducer: mockLikeKafka,
		}

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
		mockLikeKafka := NewMockKafkaProducer(ctrl)

		handler := &Handler{
			repository:        mockRepo,
			likeKafkaProducer: mockLikeKafka,
		}

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

func TestHandler_EditMaterial(t *testing.T) {
	t.Parallel()

	userUUID := uuid.New().String()
	materialUUID := uuid.New().String()

	title := "Edited Title"
	content := "Edited Content"
	description := "Edited Description"
	coverImageURL := "http://example.com/edited_cover.jpg"
	readTimeMinutes := int32(10)
	status := "draft"

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockKafka := NewMockKafkaProducer(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository:        mockRepo,
			editKafkaProducer: mockKafka,
		}

		mockRepo.EXPECT().GetMaterialOwnerUUID(gomock.Any(), materialUUID).Return(userUUID, nil)
		mockRepo.EXPECT().MaterialExists(gomock.Any(), materialUUID).Return(true, nil)

		editReq := &model.EditMaterial{
			UUID:            materialUUID,
			Title:           title,
			Content:         content,
			Description:     description,
			CoverImageURL:   coverImageURL,
			ReadTimeMinutes: readTimeMinutes,
		}

		editedMaterial := &model.Material{
			UUID:            materialUUID,
			OwnerUUID:       userUUID,
			Title:           title,
			Content:         stringPtr(content),
			Description:     description,
			CoverImageURL:   coverImageURL,
			ReadTimeMinutes: readTimeMinutes,
			Status:          status,
		}

		mockRepo.EXPECT().EditMaterial(gomock.Any(), editReq).Return(editedMaterial, nil)

		mockKafka.EXPECT().ProduceMessage(gomock.Any(), gomock.Any(), materialUUID).Return(nil)

		requestBody := api.EditMaterialIn{
			Uuid:            materialUUID,
			Title:           title,
			Content:         content,
			Description:     description,
			CoverImageUrl:   coverImageURL,
			ReadTimeMinutes: readTimeMinutes,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/edit-material", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.EditMaterial(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.EditMaterialOut
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, materialUUID, response.Material.Uuid)
		assert.Equal(t, title, response.Material.Title)
		assert.Equal(t, content, response.Material.Content)
	})

	t.Run("success_with_kafka_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockKafka := NewMockKafkaProducer(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository:        mockRepo,
			editKafkaProducer: mockKafka,
		}

		mockRepo.EXPECT().GetMaterialOwnerUUID(gomock.Any(), materialUUID).Return(userUUID, nil)
		mockRepo.EXPECT().MaterialExists(gomock.Any(), materialUUID).Return(true, nil)

		editReq := &model.EditMaterial{
			UUID:            materialUUID,
			Title:           title,
			Content:         content,
			Description:     description,
			CoverImageURL:   coverImageURL,
			ReadTimeMinutes: readTimeMinutes,
		}

		editedMaterial := &model.Material{
			UUID:            materialUUID,
			OwnerUUID:       userUUID,
			Title:           title,
			Content:         stringPtr(content),
			Description:     description,
			CoverImageURL:   coverImageURL,
			ReadTimeMinutes: readTimeMinutes,
			Status:          status,
		}

		mockRepo.EXPECT().EditMaterial(gomock.Any(), editReq).Return(editedMaterial, nil)

		mockKafka.EXPECT().ProduceMessage(gomock.Any(), gomock.Any(), materialUUID).Return(fmt.Errorf("kafka error"))

		requestBody := api.EditMaterialIn{
			Uuid:            materialUUID,
			Title:           title,
			Content:         content,
			Description:     description,
			CoverImageUrl:   coverImageURL,
			ReadTimeMinutes: readTimeMinutes,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/edit-material", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.EditMaterial(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.EditMaterialOut
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, materialUUID, response.Material.Uuid)
	})

	t.Run("invalid_json", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockKafka := NewMockKafkaProducer(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository:        mockRepo,
			editKafkaProducer: mockKafka,
		}

		req := httptest.NewRequest(http.MethodPost, "/api/materials/edit-material", strings.NewReader("invalid json"))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.EditMaterial(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Equal(t, "invalid request body", errorResp.Message)
	})

	t.Run("missing_material_uuid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockKafka := NewMockKafkaProducer(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository:        mockRepo,
			editKafkaProducer: mockKafka,
		}

		requestBody := api.EditMaterialIn{
			Uuid:            "",
			Title:           title,
			Content:         content,
			Description:     description,
			CoverImageUrl:   coverImageURL,
			ReadTimeMinutes: readTimeMinutes,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/edit-material", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.EditMaterial(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Equal(t, "material uuid is required", errorResp.Message)
	})

	t.Run("missing_title", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockKafka := NewMockKafkaProducer(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository:        mockRepo,
			editKafkaProducer: mockKafka,
		}

		requestBody := api.EditMaterialIn{
			Uuid:            materialUUID,
			Title:           " ",
			Content:         content,
			Description:     description,
			CoverImageUrl:   coverImageURL,
			ReadTimeMinutes: readTimeMinutes,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/edit-material", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.EditMaterial(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Equal(t, "title is required", errorResp.Message)
	})

	t.Run("missing_user_uuid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockKafka := NewMockKafkaProducer(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository:        mockRepo,
			editKafkaProducer: mockKafka,
		}

		requestBody := api.EditMaterialIn{
			Uuid:            materialUUID,
			Title:           title,
			Content:         content,
			Description:     description,
			CoverImageUrl:   coverImageURL,
			ReadTimeMinutes: readTimeMinutes,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/edit-material", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.EditMaterial(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Equal(t, "user UUID is required", errorResp.Message)
	})

	t.Run("error_getting_owner_uuid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockKafka := NewMockKafkaProducer(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository:        mockRepo,
			editKafkaProducer: mockKafka,
		}

		mockRepo.EXPECT().GetMaterialOwnerUUID(gomock.Any(), materialUUID).Return("", fmt.Errorf("db error"))

		requestBody := api.EditMaterialIn{
			Uuid:            materialUUID,
			Title:           title,
			Content:         content,
			Description:     description,
			CoverImageUrl:   coverImageURL,
			ReadTimeMinutes: readTimeMinutes,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/edit-material", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.EditMaterial(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Message, "failed to get owner uuid")
	})

	t.Run("user_not_owner", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockKafka := NewMockKafkaProducer(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository:        mockRepo,
			editKafkaProducer: mockKafka,
		}

		mockRepo.EXPECT().GetMaterialOwnerUUID(gomock.Any(), materialUUID).Return(uuid.New().String(), nil)

		requestBody := api.EditMaterialIn{
			Uuid:            materialUUID,
			Title:           title,
			Content:         content,
			Description:     description,
			CoverImageUrl:   coverImageURL,
			ReadTimeMinutes: readTimeMinutes,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/edit-material", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.EditMaterial(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Equal(t, "failed to edit: user is not owner", errorResp.Message)
	})

	t.Run("error_checking_existence", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockKafka := NewMockKafkaProducer(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository:        mockRepo,
			editKafkaProducer: mockKafka,
		}

		mockRepo.EXPECT().GetMaterialOwnerUUID(gomock.Any(), materialUUID).Return(userUUID, nil)
		mockRepo.EXPECT().MaterialExists(gomock.Any(), materialUUID).Return(false, fmt.Errorf("db error"))

		requestBody := api.EditMaterialIn{
			Uuid:            materialUUID,
			Title:           title,
			Content:         content,
			Description:     description,
			CoverImageUrl:   coverImageURL,
			ReadTimeMinutes: readTimeMinutes,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/edit-material", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.EditMaterial(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Message, "failed to check material existence")
	})

	t.Run("material_does_not_exist", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockKafka := NewMockKafkaProducer(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository:        mockRepo,
			editKafkaProducer: mockKafka,
		}

		mockRepo.EXPECT().GetMaterialOwnerUUID(gomock.Any(), materialUUID).Return(userUUID, nil)
		mockRepo.EXPECT().MaterialExists(gomock.Any(), materialUUID).Return(false, nil)

		requestBody := api.EditMaterialIn{
			Uuid:            materialUUID,
			Title:           title,
			Content:         content,
			Description:     description,
			CoverImageUrl:   coverImageURL,
			ReadTimeMinutes: readTimeMinutes,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/edit-material", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.EditMaterial(w, req)

		assert.Equal(t, http.StatusPreconditionFailed, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Equal(t, "material does not exist", errorResp.Message)
	})

	t.Run("repository_error_on_edit", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockKafka := NewMockKafkaProducer(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository:        mockRepo,
			editKafkaProducer: mockKafka,
		}

		mockRepo.EXPECT().GetMaterialOwnerUUID(gomock.Any(), materialUUID).Return(userUUID, nil)
		mockRepo.EXPECT().MaterialExists(gomock.Any(), materialUUID).Return(true, nil)

		editReq := &model.EditMaterial{
			UUID:            materialUUID,
			Title:           title,
			Content:         content,
			Description:     description,
			CoverImageURL:   coverImageURL,
			ReadTimeMinutes: readTimeMinutes,
		}

		mockRepo.EXPECT().EditMaterial(gomock.Any(), editReq).Return(nil, fmt.Errorf("db error"))

		requestBody := api.EditMaterialIn{
			Uuid:            materialUUID,
			Title:           title,
			Content:         content,
			Description:     description,
			CoverImageUrl:   coverImageURL,
			ReadTimeMinutes: readTimeMinutes,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/edit-material", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		reqCtx = context.WithValue(reqCtx, config.KeyUUID, userUUID)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.EditMaterial(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Message, "failed to edit material")
	})
}

func TestHandler_GetAllMaterials(t *testing.T) {
	t.Parallel()

	mockMaterials := model.MaterialList{
		{
			UUID:            uuid.New().String(),
			OwnerUUID:       uuid.New().String(),
			Title:           "Material 1",
			Content:         stringPtr("Content 1"),
			Description:     "Desc 1",
			CoverImageURL:   "url1",
			ReadTimeMinutes: 5,
			Status:          "published",
			CreatedAt:       timestamppb.Now().AsTime(),
		},
		{
			UUID:            uuid.New().String(),
			OwnerUUID:       uuid.New().String(),
			Title:           "Material 2",
			Content:         stringPtr("Content 2"),
			Description:     "Desc 2",
			CoverImageURL:   "url2",
			ReadTimeMinutes: 10,
			Status:          "published",
			CreatedAt:       timestamppb.Now().AsTime(),
		},
	}

	t.Run("success_default_params", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockRepo.EXPECT().GetAllMaterials(gomock.Any(), 0, 10).Return(&mockMaterials, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/materials/get-all-materials", nil)

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		req = req.WithContext(reqCtx)

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.GetAllMaterials(w, req, api.GetAllMaterialsParams{})

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.GetAllMaterialsOut
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response.MaterialList, len(mockMaterials))
		for i, m := range response.MaterialList {
			assert.Equal(t, mockMaterials[i].UUID, m.Uuid)
			assert.Equal(t, mockMaterials[i].Title, m.Title)
			assert.Equal(t, mockMaterials[i].Description, m.Description)
			assert.Equal(t, mockMaterials[i].CoverImageURL, m.CoverImageUrl)
			assert.Equal(t, mockMaterials[i].ReadTimeMinutes, m.ReadTimeMinutes)
			assert.Equal(t, mockMaterials[i].Status, m.Status)
			assert.Equal(t, mockMaterials[i].OwnerUUID, *m.OwnerUuid)
		}
	})

	t.Run("success_custom_params", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockRepo.EXPECT().GetAllMaterials(gomock.Any(), 10, 5).Return(&mockMaterials, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/materials/get-all-materials?page=3&limit=5", nil)

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		req = req.WithContext(reqCtx)

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.GetAllMaterials(w, req, api.GetAllMaterialsParams{})

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.GetAllMaterialsOut
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response.MaterialList, len(mockMaterials))
	})

	t.Run("invalid_page_param", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockRepo.EXPECT().GetAllMaterials(gomock.Any(), 0, 10).Return(&mockMaterials, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/materials/get-all-materials?page=0&limit=10", nil)

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		req = req.WithContext(reqCtx)

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.GetAllMaterials(w, req, api.GetAllMaterialsParams{})

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.GetAllMaterialsOut
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response.MaterialList, len(mockMaterials))
	})

	t.Run("invalid_limit_param_low", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockRepo.EXPECT().GetAllMaterials(gomock.Any(), 0, 10).Return(&mockMaterials, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/materials/get-all-materials?page=1&limit=0", nil)

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		req = req.WithContext(reqCtx)

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.GetAllMaterials(w, req, api.GetAllMaterialsParams{})

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.GetAllMaterialsOut
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response.MaterialList, len(mockMaterials))
	})

	t.Run("invalid_limit_param_high", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockRepo.EXPECT().GetAllMaterials(gomock.Any(), 0, 10).Return(&mockMaterials, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/materials/get-all-materials?page=1&limit=101", nil)

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		req = req.WithContext(reqCtx)

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.GetAllMaterials(w, req, api.GetAllMaterialsParams{})

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.GetAllMaterialsOut
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response.MaterialList, len(mockMaterials))
	})

	t.Run("repository_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockRepo.EXPECT().GetAllMaterials(gomock.Any(), 0, 10).Return(nil, fmt.Errorf("db error"))

		req := httptest.NewRequest(http.MethodGet, "/api/materials/get-all-materials", nil)

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		req = req.WithContext(reqCtx)

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.GetAllMaterials(w, req, api.GetAllMaterialsParams{})

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Message, "failed to get paginated materials")
	})
}

func TestHandler_GetMaterial(t *testing.T) {
	t.Parallel()

	materialUUID := uuid.New().String()
	mockMaterial := &model.Material{
		UUID:            materialUUID,
		OwnerUUID:       uuid.New().String(),
		Title:           "Test Title",
		Content:         stringPtr("Test Content"),
		Description:     "Test Description",
		CoverImageURL:   "http://example.com/cover.jpg",
		ReadTimeMinutes: 5,
		Status:          "published",
		CreatedAt:       timestamppb.Now().AsTime(),
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockRepo.EXPECT().GetMaterial(gomock.Any(), materialUUID).Return(mockMaterial, nil)

		requestBody := api.GetMaterialIn{
			MaterialUuid: materialUUID,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/get-material", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.GetMaterial(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.GetMaterialOut
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, mockMaterial.UUID, response.Material.Uuid)
		assert.Equal(t, mockMaterial.Title, response.Material.Title)
		assert.Equal(t, *mockMaterial.Content, response.Material.Content)
		assert.Equal(t, mockMaterial.Description, response.Material.Description)
		assert.Equal(t, mockMaterial.CoverImageURL, response.Material.CoverImageUrl)
		assert.Equal(t, mockMaterial.ReadTimeMinutes, response.Material.ReadTimeMinutes)
		assert.Equal(t, mockMaterial.Status, response.Material.Status)
		assert.Equal(t, mockMaterial.OwnerUUID, *response.Material.OwnerUuid)
	})

	t.Run("invalid_json", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		req := httptest.NewRequest(http.MethodPost, "/api/materials/get-material", bytes.NewReader([]byte("invalid json")))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.GetMaterial(w, req)

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

		requestBody := api.GetMaterialIn{
			MaterialUuid: "",
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/get-material", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.GetMaterial(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Message, "material uuid is required")
	})

	t.Run("repository_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockRepo.EXPECT().GetMaterial(gomock.Any(), materialUUID).Return(nil, fmt.Errorf("db error"))

		requestBody := api.GetMaterialIn{
			MaterialUuid: materialUUID,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/get-material", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.GetMaterial(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Message, "failed to get material")
	})

	t.Run("material_not_found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := NewMockDBRepo(ctrl)
		mockLogger := logger_lib.NewMockLoggerInterface(ctrl)

		handler := &Handler{
			repository: mockRepo,
		}

		mockRepo.EXPECT().GetMaterial(gomock.Any(), materialUUID).Return(nil, fmt.Errorf("material doesn't exist"))

		requestBody := api.GetMaterialIn{
			MaterialUuid: materialUUID,
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/materials/get-material", bytes.NewReader(bodyBytes))

		reqCtx := req.Context()
		reqCtx = context.WithValue(reqCtx, config.KeyLogger, mockLogger)
		req = req.WithContext(reqCtx)

		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.GetMaterial(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var errorResp api.Error
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp.Message, "material does not exist")
	})
}

func stringPtr(s string) *string {
	return &s
}

//func int32Ptr(i int32) *int32 {
//	return &i
//}
