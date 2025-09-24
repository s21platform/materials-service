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
)

type Handler struct {
	repository DBRepo
}

func New(repo DBRepo) *Handler {
	return &Handler{
		repository: repo,
	}
}

func (h *Handler) SaveDraftMaterial(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req api.SaveDraftMaterialIn
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to decode request")
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
		Content:         *req.Content,
		Description:     *req.Description,
		CoverImageURL:   *req.CoverImageUrl,
		ReadTimeMinutes: *req.ReadTimeMinutes,
	}

	respUUID, err := h.repository.SaveDraftMaterial(ctx, userUUID, saveReq)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to save draft material")
		h.writeError(w, fmt.Sprintf("failed to save draft material: %v", err), http.StatusInternalServerError)
		return
	}

	response := api.SaveDraftMaterialOut{
		Uuid: respUUID,
	}

	h.writeJSON(w, response, http.StatusOK)
}

func (h *Handler) PublishMaterial(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req api.PublishMaterialIn
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to decode request")
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

	ctx = logger_lib.WithUserUuid(ctx, userUUID)

	materialOwnerUUID, err := h.repository.GetMaterialOwnerUUID(ctx, req.Uuid)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to get owner uuid")
		h.writeError(w, fmt.Sprintf("failed to get owner uuid: %v", err), http.StatusInternalServerError)
		return
	}

	if materialOwnerUUID != userUUID {
		logger_lib.Error(ctx, "failed to publish: user is not owner")
		h.writeError(w, "failed to publish: user is not owner", http.StatusForbidden)
		return
	}

	exists, err := h.repository.MaterialExists(ctx, req.Uuid)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to check material existence")
		h.writeError(w, fmt.Sprintf("failed to check material existence: %v", err), http.StatusInternalServerError)
		return
	}

	if !exists {
		logger_lib.Error(ctx, "material does not exist")
		h.writeError(w, "material does not exist", http.StatusPreconditionFailed)
		return
	}

	publishedMaterial, err := h.repository.PublishMaterial(ctx, req.Uuid)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to publish material")
		h.writeError(w, fmt.Sprintf("failed to publish material: %v", err), http.StatusInternalServerError)
		return
	}

	response := api.PublishMaterialOut{
		Material: api.Material{
			Uuid:            publishedMaterial.UUID,
			OwnerUuid:       &publishedMaterial.OwnerUUID,
			Title:           publishedMaterial.Title,
			Content:         publishedMaterial.Content,
			Description:     &publishedMaterial.Description,
			CoverImageUrl:   &publishedMaterial.CoverImageURL,
			ReadTimeMinutes: &publishedMaterial.ReadTimeMinutes,
			Status:          &publishedMaterial.Status,
		},
	}

	h.writeJSON(w, response, http.StatusOK)
}

func (h *Handler) ToggleLike(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req api.ToggleLikeIn
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to decode request")
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

	ctx = logger_lib.WithUserUuid(ctx, userUUID)

	var isLiked bool
	var likesCount int32
	err := tx.TxExecute(ctx, func(ctx context.Context) error {
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
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to toggle like")
		h.writeError(w, fmt.Sprintf("failed to toggle like: %v", err), http.StatusInternalServerError)
		return
	}

	response := api.ToggleLikeOut{
		IsLiked:    !isLiked,
		LikesCount: likesCount,
	}

	h.writeJSON(w, response, http.StatusOK)
}

// ----------------------------- helpers -----------------------------

func (h *Handler) writeJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger_lib.Error(logger_lib.WithError(context.Background(), err), "failed to encode response")
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
		logger_lib.Error(logger_lib.WithError(context.Background(), err), "failed to encode error response")
		http.Error(w, "failed to encode error response", http.StatusInternalServerError)
	}
}
