package avatar

import (
	"context"
	"encoding/json"

	logger_lib "github.com/s21platform/logger-lib"

	"github.com/s21platform/avatar-service/pkg/avatar"
)

type Handler struct {
	repository DBRepo
}

func New(repo DBRepo) *Handler {
	return &Handler{repository: repo}
}

func convertMessage(bMessage []byte, target interface{}) error {
	err := json.Unmarshal(bMessage, target)
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) Handler(ctx context.Context, in []byte) error {
	var msg avatar.NewAvatarRegister
	err := convertMessage(in, &msg)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to convert message")
		return err
	}

	ctx = logger_lib.WithUserUuid(ctx, msg.Uuid)

	err = h.repository.AvatarLinkUpdate(ctx, msg.Uuid, msg.Link)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to update avatar link")
		return err
	}

	return nil
}
