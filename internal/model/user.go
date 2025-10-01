package model

type User struct {
	Uuid       string `db:"uuid"`
	Nickname   string `db:"nickname"`
	AvatarLink string `db:"avatar_link"`
	Name       string `db:"name"`
	Surname    string `db:"surname"`
}
