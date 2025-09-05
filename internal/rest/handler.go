package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	logger_lib "github.com/s21platform/logger-lib"
	"github.com/s21platform/materials-service/internal/config"
	api "github.com/s21platform/materials-service/internal/generated"
	"github.com/s21platform/materials-service/internal/service"
	"github.com/s21platform/materials-service/pkg/materials"
)

type Handler struct {
	repository DBRepo
	service    *service.Service
}

func New(
	repo service.DBRepo,
	svc *service.Service) *Handler {
	return &Handler{
		repository: repo,
		service:    svc,
	}
}

func (h *Handler) SaveDraftMaterial(w http.ResponseWriter, r *http.Request, params api.SaveDraftMaterialParams) {
	logger := logger_lib.FromContext(r.Context(), config.KeyLogger)
	logger.AddFuncName("CreateStream")

	var req api.SaveDraftMaterialIn
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error(fmt.Sprintf("failed to decode request: %v", err))
		h.writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userUUID, ok := r.Context().Value(config.KeyUUID).(string)
	if !ok || userUUID == "" {
		logger.Error("failed to get user UUID")
		h.writeError(w, "user UUID is required", http.StatusUnauthorized)
		return
	}

	if userUUID != params.XUserID {
		logger.Error("user UUID mismatch with X-User-ID header")
		h.writeError(w, "user UUID mismatch", http.StatusUnauthorized)
		return
	}

	saveReq := &materials.SaveDraftMaterialIn{
		Title:           req.Title,
		Content:         *req.Content,
		Description:     *req.Description,
		CoverImageUrl:   *req.CoverImageUrl,
		ReadTimeMinutes: *req.ReadTimeMinutes,
	}

	ctx := r.Context()
	resp, err := h.service.SaveDraftMaterial(ctx, saveReq)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to save draft material: %v", err))
		h.writeError(w, fmt.Sprintf("failed to save draft material: %v", err), http.StatusInternalServerError)
		return
	}

	response := api.SaveDraftMaterialOut{
		Uuid: resp.Uuid,
	}

	h.writeJSON(w, response, http.StatusOK)
}

// ----------------------------- helpers -----------------------------

func (h *Handler) writeJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger := logger_lib.FromContext(context.Background(), config.KeyLogger)
		logger.Error(fmt.Sprintf("failed to encode response: %v", err))
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
		logger := logger_lib.FromContext(context.Background(), config.KeyLogger)
		logger.Error(fmt.Sprintf("failed to encode error response: %v", err))
		http.Error(w, "failed to encode error response", http.StatusInternalServerError)
	}
}
