package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Импорт драйвера PostgreSQL

	"github.com/s21platform/materials-service/internal/config"
	"github.com/s21platform/materials-service/internal/model"
)

type Repository struct {
	connection *sqlx.DB
}

func New(cfg *config.Config) *Repository {
	conStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Database, cfg.Postgres.Host, cfg.Postgres.Port)

	conn, err := sqlx.Connect("postgres", conStr)
	if err != nil {
		log.Fatal("failed to connect: ", err)
	}

	return &Repository{
		connection: conn,
	}
}

func (r *Repository) Close() {
	_ = r.connection.Close()
}

func (r *Repository) SaveDraftMaterial(ctx context.Context, ownerUUID string, material *model.SaveDraftMaterial) (string, error) {
	var uuid string

	query, args, err := sq.
		Insert("materials").
		Columns("owner_uuid", "title", "cover_image_url", "description", "content", "read_time_minutes").
		Values(ownerUUID, material.Title, material.CoverImageURL, material.Description, material.Content, material.ReadTimeMinutes).
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

func (r *Repository) GetMaterial(ctx context.Context, uuid string) (*model.Material, error) {
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
		Where(sq.Eq{"uuid": uuid}).
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

func (r *Repository) GetAllMaterials(ctx context.Context) (*model.MaterialList, error) {
	var materials model.MaterialList

	query, args, err := sq.
		Select(
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
		OrderBy("created_at DESC").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	err = r.connection.SelectContext(ctx, &materials, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch materials: %w", err)
	}

	return &materials, nil
}

func (r *Repository) EditMaterial(ctx context.Context, material *model.EditMaterial) (*model.Material, error) {
	var updatedMaterial model.Material
	query, args, err := sq.
		Update("materials").
		Set("title", material.Title).
		Set("cover_image_url", material.CoverImageURL).
		Set("description", material.Description).
		Set("content", material.Content).
		Set("read_time_minutes", material.ReadTimeMinutes).
		Set("edited_at", time.Now()).
		Where(sq.Eq{"uuid": material.UUID}).
		PlaceholderFormat(sq.Dollar).
		Suffix("RETURNING uuid, owner_uuid, title, cover_image_url, description, content, read_time_minutes, status, created_at, edited_at, published_at, archived_at, deleted_at, likes_count").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build update query: %v", err)
	}

	err = r.connection.GetContext(ctx, &updatedMaterial, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update material: %v", err)
	}

	return &updatedMaterial, nil
}

func (r *Repository) GetMaterialOwnerUUID(ctx context.Context, uuid string) (string, error) {
	var ownerUUID string

	query, args, err := sq.
		Select("owner_uuid").
		From("materials").
		Where(sq.Eq{"uuid": uuid}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("failed to build sql query: %v", err)
	}

	err = r.connection.GetContext(ctx, &ownerUUID, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("material doesn't exist")
		}
		return "", fmt.Errorf("failed to get owner uuid: %v", err)
	}

	return ownerUUID, nil
}

func (r *Repository) DeleteMaterial(ctx context.Context, uuid string) (int64, error) {
	query, args, err := sq.
		Update("materials").
		Set("deleted_at", time.Now()).
		Where(sq.Eq{"uuid": uuid, "deleted_at": nil}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build sql query: %v", err)
	}

	res, err := r.connection.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute query: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to check rows affected: %v", err)
	}

	return rowsAffected, nil
}

func (r *Repository) PublishMaterial(ctx context.Context, uuid string) (*model.Material, error) {
	var updatedMaterial model.Material

	query, args, err := sq.
		Update("materials").
		Set("status", "published").
		Set("published_at", time.Now()).
		Where(sq.Eq{"uuid": uuid}).
		PlaceholderFormat(sq.Dollar).
		Suffix("RETURNING uuid, owner_uuid, title, cover_image_url, description, content, read_time_minutes, status, created_at, edited_at, published_at, archived_at, deleted_at, likes_count").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build update query: %v", err)
	}

	err = r.connection.GetContext(ctx, &updatedMaterial, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute update query: %v", err)
	}

	return &updatedMaterial, nil
}

func (r *Repository) MaterialExists(ctx context.Context, materialUUID string) (bool, error) {
	var exists bool

	subQuery := sq.
		Select("1").
		From("materials").
		Where(sq.Eq{"uuid": materialUUID}).
		Where("deleted_at IS NULL")

	query := sq.
		Select().
		Column(sq.Expr("EXISTS(?) AS exists", subQuery)).
		PlaceholderFormat(sq.Dollar)

	querySQL, args, err := query.ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build sql query: %v", err)
	}

	err = r.connection.GetContext(ctx, &exists, querySQL, args...)
	if err != nil {
		return false, fmt.Errorf("failed to check material existence: %v", err)
	}

	return exists, nil
}
