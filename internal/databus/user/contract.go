package user

import "context"

type DBRepo interface {
	CreateUser(ctx context.Context, userUUID, nickname, avatarLink, name, surname string) error
}
