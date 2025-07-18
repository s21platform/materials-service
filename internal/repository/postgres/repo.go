package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/s21platform/materials-service/internal/config"
	"github.com/s21platform/materials-service/pkg/materials"
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

func (r *MaterialRepository) GetMaterial(ctx context.Context, Uuid string) (*materials.GetMaterialOut, error) {
	query, args, err := sq.Select("uuid",
		"owner_uuid AS OwnerUuid",
		"title",
		"cover_image_url AS CoverImageUrl",
		"description",
		"content",
		"read_time_minutes AS ReadTimeMinutes",
	).
		From("materials").
		Where(sq.Eq{"uuid": Uuid}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sql query: %v", err)
	}

	var materialResponse materials.GetMaterialOut
	err = r.connection.GetContext(ctx, &materialResponse, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("material not found: %v", err)
		}
		return nil, fmt.Errorf("failed to get material: %v", err)
	}

	return &materialResponse, nil
}
