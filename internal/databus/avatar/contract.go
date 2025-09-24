package avatar

import "context"

type DBRepo interface {
	AvatarLinkUpdate(ctx context.Context, userUUID string, avatarLink string) error
}
