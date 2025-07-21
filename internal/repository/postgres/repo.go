package postgres

import (
	"context"
	"fmt"
	"log"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Импорт драйвера PostgreSQL

	"github.com/s21platform/materials-service/internal/config"
	"github.com/s21platform/materials-service/pkg/materials"
)

type Repository struct {
	connection *sqlx.DB
}

func New(cfg *config.Config) *Repository {
	conStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Database, cfg.Postgres.Host, cfg.Postgres.Port)

	conn, err := sqlx.Connect("postgres", conStr)
	if err != nil {
		log.Fatal("error connect: ", err)
	}

	return &Repository{
		connection: conn,
	}
}

func (r *Repository) Close() {
	_ = r.connection.Close()
}

func (r *Repository) CreateMaterial(ctx context.Context, ownerUUID string, in *materials.CreateMaterialIn) (string, error) {
	var uuid string

	query, args, err := sq.
		Insert("materials").
		Columns("owner_uuid", "title", "cover_image_url", "description", "content", "read_time_minutes").
		Values(ownerUUID, in.Title, in.CoverImageUrl, in.Description, in.Content, in.ReadTimeMinutes).
		PlaceholderFormat(sq.Dollar).
		Suffix("RETURNING uuid").
		ToSql()

	if err != nil {
		return "", fmt.Errorf("failed to build SQL query: %w", err)
	}

	err = r.connection.GetContext(ctx, &uuid, query, args...)
	if err != nil {
		return "", fmt.Errorf("failed to execute query: %w", err)
	}

	return uuid, nil
}
