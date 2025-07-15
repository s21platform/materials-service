package postgres

import (
	"context"
	_ "database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/s21platform/materials-service/internal/config"
	"github.com/s21platform/materials-service/pkg/materials"
	_ "github.com/s21platform/materials-service/pkg/materials"
	"log"
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

func (r *Repository) CreateMaterial(ctx context.Context, uuid string, in *materials.CreateMaterialIn) (*materials.CreateMaterialOut, error) {
	query, args, err := sq.
		Insert("materials").
		Columns("owner_uuid", "title", "description", "content", "read_time_minutes").
		Values(in.OwnerUuid, in.Title, in.Description, in.Content, in.ReadTimeMinutes).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.connection.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to insert material: %w", err)
	}

	return &materials.CreateMaterialOut{
		MaterialUuid: uuid,
	}, nil
}
