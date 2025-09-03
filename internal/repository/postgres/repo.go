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

func (r *Repository) Conn() *sqlx.DB {
	return r.connection
}

func (r *Repository) CreateMaterial(ctx context.Context, ownerUUID string, material *model.CreateMaterial) (string, error) {
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

func (r *Repository) CheckLike(ctx context.Context, materialUUID string, userUUID string, tx *sqlx.Tx) (bool, error) {
	var exists bool

	query, _, err := sq.
		Select("EXISTS (SELECT 1 FROM material_likes WHERE material_uuid = ? AND user_uuid = ?)").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build sql query: %w", err)
	}

	err = tx.GetContext(ctx, &exists, query, materialUUID, userUUID)
	if err != nil {
		return false, fmt.Errorf("failed to check like: %w", err)
	}

	return exists, nil
}

func (r *Repository) AddLike(ctx context.Context, materialUUID string, userUUID string, tx *sqlx.Tx) (bool, error) {
	query, args, err := sq.
		Insert("material_likes").
		Columns("uuid", "material_uuid", "user_uuid", "created_at").
		Values(sq.Expr("gen_random_uuid()"), materialUUID, userUUID, time.Now()).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build sql query: %w", err)
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return false, fmt.Errorf("failed to add like: %w", err)
	}

	return true, nil
}

func (r *Repository) RemoveLike(ctx context.Context, materialUUID string, userUUID string, tx *sqlx.Tx) (bool, error) {
	query, args, err := sq.
		Delete("material_likes").
		Where(sq.Eq{"material_uuid": materialUUID, "user_uuid": userUUID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build sql query: %w", err)
	}

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return false, fmt.Errorf("failed to remove like: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected > 0, nil
}

func (r *Repository) GetLikesCount(ctx context.Context, materialUUID string, tx *sqlx.Tx) (int32, error) {
	var likesCount int32

	query, args, err := sq.
		Select("COUNT(material_uuid)").
		From("material_likes").
		Where(sq.Eq{"material_uuid": materialUUID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build sql query: %w", err)
	}

	err = tx.GetContext(ctx, &likesCount, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to get likes count: %w", err)
	}

	return likesCount, nil
}

func (r *Repository) UpdateLikesCount(ctx context.Context, materialUUID string, likesCount int32, tx *sqlx.Tx) (int32, error) {
	query, args, err := sq.
		Update("materials").
		Set("likes_count", likesCount).
		Where(sq.Eq{"uuid": materialUUID}).
		Suffix("RETURNING likes_count").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build sql query: %w", err)
	}

	var updatedLikesCount int32
	err = tx.GetContext(ctx, &updatedLikesCount, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to update likes count: %w", err)
	}

	return updatedLikesCount, nil
}
