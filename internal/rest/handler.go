package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	logger_lib "github.com/s21platform/logger-lib"

	"github.com/s21platform/materials-service/internal/config"
	api "github.com/s21platform/materials-service/internal/generated"
	"github.com/s21platform/materials-service/internal/model"
	"github.com/s21platform/materials-service/internal/pkg/tx"
	proto "github.com/s21platform/materials-service/pkg/materials"
)

type Handler struct {
	repository          DBRepo
	createKafkaProducer KafkaProducer
	likeKafkaProducer   KafkaProducer
}

func New(repo DBRepo, createKafkaProducer KafkaProducer, likeKafkaProducer KafkaProducer) *Handler {
	return &Handler{
		repository:          repo,
		createKafkaProducer: createKafkaProducer,
		likeKafkaProducer:   likeKafkaProducer,
	}
}

func (h *Handler) SaveDraftMaterial(w http.ResponseWriter, r *http.Request) {
	ctx := logger_lib.WithField(r.Context(), "func_name", "SaveDraftMaterial")

	var req api.SaveDraftMaterialIn
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), fmt.Sprintf("failed to decode request: %v", err))
		h.writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userUUID, ok := r.Context().Value(config.KeyUUID).(string)
	if !ok || userUUID == "" {
		logger_lib.Error(ctx, "failed to get user UUID")
		h.writeError(w, "user UUID is required", http.StatusUnauthorized)
		return
	}

	saveReq := &model.SaveDraftMaterial{
		Title:           req.Title,
		Content:         req.Content,
		Description:     req.Description,
		CoverImageURL:   req.CoverImageUrl,
		ReadTimeMinutes: req.ReadTimeMinutes,
	}

	respUUID, err := h.repository.SaveDraftMaterial(r.Context(), userUUID, saveReq)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), fmt.Sprintf("failed to save draft material: %v", err))
		h.writeError(w, fmt.Sprintf("failed to save draft material: %v", err), http.StatusInternalServerError)
		return
	}

	response := api.SaveDraftMaterialOut{
		Uuid: respUUID,
	}

	h.writeJSON(w, response, http.StatusOK)
}

func (h *Handler) PublishMaterial(w http.ResponseWriter, r *http.Request) {
	ctx := logger_lib.WithField(r.Context(), "func_name", "PublishMaterial")

	var req api.PublishMaterialIn
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), fmt.Sprintf("failed to decode request: %v", err))
		h.writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Uuid == "" {
		logger_lib.Error(ctx, "material uuid is required")
		h.writeError(w, "material uuid is required", http.StatusBadRequest)
		return
	}

	userUUID, ok := r.Context().Value(config.KeyUUID).(string)
	if !ok || userUUID == "" {
		logger_lib.Error(ctx, "failed to get user UUID")
		h.writeError(w, "user UUID is required", http.StatusUnauthorized)
		return
	}

	materialOwnerUUID, err := h.repository.GetMaterialOwnerUUID(r.Context(), req.Uuid)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), fmt.Sprintf("failed to get owner uuid: %v", err))
		h.writeError(w, fmt.Sprintf("failed to get owner uuid: %v", err), http.StatusInternalServerError)
		return
	}

	if materialOwnerUUID != userUUID {
		logger_lib.Error(ctx, "failed to publish: user is not owner")
		h.writeError(w, "failed to publish: user is not owner", http.StatusForbidden)
		return
	}

	exists, err := h.repository.MaterialExists(r.Context(), req.Uuid)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), fmt.Sprintf("failed to check material existence: %v", err))
		h.writeError(w, fmt.Sprintf("failed to check material existence: %v", err), http.StatusInternalServerError)
		return
	}

	if !exists {
		logger_lib.Error(ctx, "material does not exist")
		h.writeError(w, "material does not exist", http.StatusPreconditionFailed)
		return
	}

	publishedMaterial, err := h.repository.PublishMaterial(r.Context(), req.Uuid)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), fmt.Sprintf("failed to publish material: %v", err))
		h.writeError(w, fmt.Sprintf("failed to publish material: %v", err), http.StatusInternalServerError)
		return
	}

	response := api.PublishMaterialOut{
		Material: api.Material{
			Uuid:            publishedMaterial.UUID,
			OwnerUuid:       &publishedMaterial.OwnerUUID,
			Title:           publishedMaterial.Title,
			Content:         *publishedMaterial.Content,
			Description:     publishedMaterial.Description,
			CoverImageUrl:   publishedMaterial.CoverImageURL,
			ReadTimeMinutes: publishedMaterial.ReadTimeMinutes,
			Status:          publishedMaterial.Status,
		},
	}

	createMaterial := &proto.CreatedMaterial{
		Material: &proto.Material{
			Uuid:            publishedMaterial.UUID,
			OwnerUuid:       publishedMaterial.OwnerUUID,
			Title:           publishedMaterial.Title,
			Content:         *publishedMaterial.Content,
			Description:     publishedMaterial.Description,
			CoverImageUrl:   publishedMaterial.CoverImageURL,
			ReadTimeMinutes: publishedMaterial.ReadTimeMinutes,
			Status:          publishedMaterial.Status,
		},
	}

	err = h.createKafkaProducer.ProduceMessage(r.Context(), createMaterial, materialOwnerUUID)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to produce message")
	}

	h.writeJSON(w, response, http.StatusOK)
}

func (h *Handler) ToggleLike(w http.ResponseWriter, r *http.Request) {
	ctx := logger_lib.WithField(r.Context(), "func_name", "ToggleLike")

	var req api.ToggleLikeIn
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), fmt.Sprintf("failed to decode request: %v", err))
		h.writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.MaterialUuid == "" {
		logger_lib.Error(ctx, "material uuid is required")
		h.writeError(w, "material uuid is required", http.StatusBadRequest)
		return
	}

	userUUID, ok := r.Context().Value(config.KeyUUID).(string)
	if !ok || userUUID == "" {
		logger_lib.Error(ctx, "failed to get user UUID")
		h.writeError(w, "user UUID is required", http.StatusUnauthorized)
		return
	}

	var isLiked bool
	var likesCount int32
	err := tx.TxExecute(r.Context(), func(ctx context.Context) error {
		var err error
		isLiked, err = h.repository.CheckLike(ctx, req.MaterialUuid, userUUID)
		if err != nil {
			return fmt.Errorf("failed to check like: %v", err)
		}

		if isLiked {
			err = h.repository.RemoveLike(ctx, req.MaterialUuid, userUUID)
			if err != nil {
				return fmt.Errorf("failed to remove like: %v", err)
			}
		} else {
			err = h.repository.AddLike(ctx, req.MaterialUuid, userUUID)
			if err != nil {
				return fmt.Errorf("failed to add like: %v", err)
			}
		}

		likesCount, err = h.repository.GetLikesCount(ctx, req.MaterialUuid)
		if err != nil {
			return fmt.Errorf("failed to get likes count: %v", err)
		}

		err = h.repository.UpdateLikesCount(ctx, req.MaterialUuid, likesCount)
		if err != nil {
			return fmt.Errorf("failed to update likes count: %v", err)
		}

		return nil
	})
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), fmt.Sprintf("failed to toggle like: %v", err))
		h.writeError(w, fmt.Sprintf("failed to toggle like: %v", err), http.StatusInternalServerError)
		return
	}

	response := api.ToggleLikeOut{
		IsLiked:    !isLiked,
		LikesCount: likesCount,
	}

	likeMsg := &proto.ToggleLikeMessage{
		MaterialUuid: req.MaterialUuid,
		IsLiked:      isLiked,
		LikesCount:   likesCount,
	}

	err = h.likeKafkaProducer.ProduceMessage(r.Context(), likeMsg, req.MaterialUuid)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to produce message")
	}

	h.writeJSON(w, response, http.StatusOK)
}

// ----------------------------- helpers -----------------------------

func (h *Handler) writeJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger_lib.Error(context.Background(), fmt.Sprintf("failed to encode response: %v", err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(api.Error{
		Code:    statusCode,
		Message: message,
	})
	if err != nil {
		logger_lib.Error(context.Background(), fmt.Sprintf("failed to encode error response: %v", err))
		http.Error(w, "failed to encode error response", http.StatusInternalServerError)
	}
}
