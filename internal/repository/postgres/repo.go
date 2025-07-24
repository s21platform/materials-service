package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Импорт драйвера PostgreSQL

	"github.com/s21platform/materials-service/internal/config"
	"github.com/s21platform/materials-service/internal/model"
)

type MaterialRepository struct {
	connection *sqlx.DB
}

func New(cfg *config.Config) *MaterialRepository {
	conStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Database, cfg.Postgres.Host, cfg.Postgres.Port)

	conn, err := sqlx.Connect("postgres", conStr)
	if err != nil {
		log.Fatal("error connect: ", err)
	}

	return &MaterialRepository{
		connection: conn,
	}
}

func (r *MaterialRepository) Close() {
	_ = r.connection.Close()
}

func (r *MaterialRepository) GetMaterial(ctx context.Context, Uuid string) (*model.Material, error) {
	var material model.Material

	query, args, err := sq.Select(
		"uuid",
		"owner_uuid",
		"title",
		"cover_image_url",
		"description",
		"content",
		"read_time_minutes",
		"status",
		"created_at",
		"edited_at",
		"published_at",
		"archived_at",
		"deleted_at",
		"likes_count",
	).
		From("materials").
		Where(sq.Eq{"uuid": Uuid}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sql query: %v", err)
	}

	err = r.connection.GetContext(ctx, &material, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("material doesn't exist")
		}
		return nil, fmt.Errorf("failed to get material: %v", err)
	}

	return &material, nil
}
