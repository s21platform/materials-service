package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	logger_lib "github.com/s21platform/logger-lib"
	"google.golang.org/protobuf/types/known/timestamppb"

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
	editKafkaProducer   KafkaProducer
}

func New(repo DBRepo, createKafkaProducer, likeKafkaProducer, editKafkaProducer KafkaProducer) *Handler {
	return &Handler{
		repository:          repo,
		createKafkaProducer: createKafkaProducer,
		likeKafkaProducer:   likeKafkaProducer,
		editKafkaProducer:   editKafkaProducer,
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

	likeMsg := &proto.ToggleLikeMessage{
		MaterialUuid: req.MaterialUuid,
		IsLiked:      !isLiked,
		LikesCount:   likesCount,
	}

	err = h.likeKafkaProducer.ProduceMessage(r.Context(), likeMsg, req.MaterialUuid)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to produce message")
	}

	response := api.ToggleLikeOut{
		IsLiked:    !isLiked,
		LikesCount: likesCount,
	}

	h.writeJSON(w, response, http.StatusOK)
}

func (h *Handler) EditMaterial(w http.ResponseWriter, r *http.Request) {
	ctx := logger_lib.WithField(r.Context(), "func_name", "EditMaterial")

	var req api.EditMaterialIn
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

	if strings.TrimSpace(req.Title) == "" {
		logger_lib.Error(ctx, "title is required")
		h.writeError(w, "title is required", http.StatusBadRequest)
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
		logger_lib.Error(ctx, "failed to edit: user is not owner")
		h.writeError(w, "failed to edit: user is not owner", http.StatusForbidden)
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

	editReq := &model.EditMaterial{
		UUID:            req.Uuid,
		Title:           req.Title,
		Content:         req.Content,
		Description:     req.Description,
		CoverImageURL:   req.CoverImageUrl,
		ReadTimeMinutes: req.ReadTimeMinutes,
	}

	editedMaterial, err := h.repository.EditMaterial(r.Context(), editReq)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), fmt.Sprintf("failed to edit material: %v", err))
		h.writeError(w, fmt.Sprintf("failed to edit material: %v", err), http.StatusInternalServerError)
		return
	}

	response := api.EditMaterialOut{
		Material: api.Material{
			Uuid:            editedMaterial.UUID,
			OwnerUuid:       &editedMaterial.OwnerUUID,
			Title:           editedMaterial.Title,
			Content:         *editedMaterial.Content,
			Description:     editedMaterial.Description,
			CoverImageUrl:   editedMaterial.CoverImageURL,
			ReadTimeMinutes: editedMaterial.ReadTimeMinutes,
			Status:          editedMaterial.Status,
		},
	}

	editMsg := &proto.EditMaterialMessage{
		Uuid:      req.Uuid,
		OwnerUuid: materialOwnerUUID,
		Title:     req.Title,
		EditedAt:  timestamppb.New(time.Now()),
	}

	err = h.editKafkaProducer.ProduceMessage(r.Context(), editMsg, req.Uuid)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to produce message")
	}

	h.writeJSON(w, response, http.StatusOK)
}

func (h *Handler) GetAllMaterials(w http.ResponseWriter, r *http.Request, params api.GetAllMaterialsParams) {
	ctx := logger_lib.WithField(r.Context(), "func_name", "GetAllMaterials")

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		logger_lib.Warn(ctx, "invalid page parameter, using default 1")
		page = 1
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		logger_lib.Warn(ctx, "invalid limit parameter, using default 10")
		limit = 10
	}
	offset := (page - 1) * limit

	paginatedMaterials, err := h.repository.GetAllMaterials(r.Context(), offset, limit)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), fmt.Sprintf("failed to get paginated materials: %v", err))
		h.writeError(w, fmt.Sprintf("failed to get paginated materials: %v", err), http.StatusInternalServerError)
		return
	}

	response := api.GetAllMaterialsOut{
		MaterialList: func(materialsList *model.MaterialList) []api.Material {
			var apiList []api.Material
			for _, m := range *materialsList {
				apiList = append(apiList, api.Material{
					Uuid:            m.UUID,
					OwnerUuid:       &m.OwnerUUID,
					Title:           m.Title,
					Description:     m.Description,
					CoverImageUrl:   m.CoverImageURL,
					ReadTimeMinutes: m.ReadTimeMinutes,
					Status:          m.Status,
				})
			}
			return apiList
		}(paginatedMaterials),
	}

	h.writeJSON(w, response, http.StatusOK)
}

func (h *Handler) GetMaterial(w http.ResponseWriter, r *http.Request) {
	ctx := logger_lib.WithField(r.Context(), "func_name", "GetMaterial")

	var req api.GetMaterialIn
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

	material, err := h.repository.GetMaterial(r.Context(), req.MaterialUuid)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to get material")
		if strings.Contains(err.Error(), "material doesn't exist") {
			h.writeError(w, "material does not exist", http.StatusNotFound)
		} else {
			h.writeError(w, fmt.Sprintf("failed to get material: %v", err), http.StatusInternalServerError)
		}
		return
	}

	response := api.GetMaterialOut{
		Material: api.Material{
			Uuid:            material.UUID,
			OwnerUuid:       &material.OwnerUUID,
			Title:           material.Title,
			Content:         *material.Content,
			Description:     material.Description,
			CoverImageUrl:   material.CoverImageURL,
			ReadTimeMinutes: material.ReadTimeMinutes,
			Status:          material.Status,
		},
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
