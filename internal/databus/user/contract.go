package user

import (
	"context"

	"github.com/s21platform/materials-service/internal/model"
)

type DBRepo interface {
	UpdateUserNickname(ctx context.Context, userUUID, newNickname string) error
	CreateUser(ctx context.Context, user model.User) error
}
