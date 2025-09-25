package user

import (
	"context"

	"github.com/s21platform/materials-service/internal/model"
)

type DBRepo interface {
	CreateUser(ctx context.Context, user model.User) error
}
