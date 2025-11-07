package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/s21platform/materials-service/internal/config"
	"github.com/s21platform/materials-service/internal/model"
)

const (
	prefix = "material:"
)

type Repository struct {
	conn *redis.Client
}

func New(cfg *config.Config) *Repository {
	redisAddr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)

	rdb := redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		DB:           0,
		MinIdleConns: 2,
		Protocol:     2,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatal(err)
	}

	return &Repository{conn: rdb}
}

func (r *Repository) Close() {
	_ = r.conn.Close()
}

func (r *Repository) SetMaterial(ctx context.Context, material *model.Material, ttl time.Duration) error {
	key := prefix + material.UUID

	data := map[string]interface{}{
		"uuid":              material.UUID,
		"owner_uuid":        material.OwnerUUID,
		"title":             material.Title,
		"cover_image_url":   material.CoverImageURL,
		"description":       material.Description,
		"read_time_minutes": material.ReadTimeMinutes,
		"status":            material.Status,
		"created_at":        material.CreatedAt.Format(time.RFC3339),
		"likes_count":       material.LikesCount,
	}

	if material.Content != nil {
		data["content"] = *material.Content
	}
	if material.EditedAt != nil {
		data["edited_at"] = material.EditedAt.Format(time.RFC3339)
	}
	if material.PublishedAt != nil {
		data["published_at"] = material.PublishedAt.Format(time.RFC3339)
	}
	if material.ArchivedAt != nil {
		data["archived_at"] = material.ArchivedAt.Format(time.RFC3339)
	}
	if material.DeletedAt != nil {
		data["deleted_at"] = material.DeletedAt.Format(time.RFC3339)
	}

	if err := r.conn.HSet(ctx, key, data).Err(); err != nil {
		return err
	}

	if ttl > 0 {
		r.conn.Expire(ctx, key, ttl)
	}

	return nil
}

func (r *Repository) GetMaterial(ctx context.Context, uuid string) (*model.Material, error) {
	key := prefix + uuid

	data, err := r.conn.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, redis.Nil
	}

	parseTime := func(s string) (*time.Time, error) {
		if s == "" {
			return nil, nil
		}
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return nil, err
		}
		return &t, nil
	}

	createdAt, _ := time.Parse(time.RFC3339, data["created_at"])

	material := &model.Material{
		UUID:            data["uuid"],
		OwnerUUID:       data["owner_uuid"],
		Title:           data["title"],
		CoverImageURL:   data["cover_image_url"],
		Description:     data["description"],
		ReadTimeMinutes: parseInt32(data["read_time_minutes"]),
		Status:          data["status"],
		CreatedAt:       createdAt,
		LikesCount:      parseInt32(data["likes_count"]),
	}

	if content, ok := data["content"]; ok && content != "" {
		material.Content = &content
	}
	if editedAtStr, ok := data["edited_at"]; ok && editedAtStr != "" {
		if t, err := parseTime(editedAtStr); err == nil && t != nil {
			material.EditedAt = t
		}
	}
	if publishedAtStr, ok := data["published_at"]; ok && publishedAtStr != "" {
		if t, err := parseTime(publishedAtStr); err == nil && t != nil {
			material.PublishedAt = t
		}
	}
	if archivedAtStr, ok := data["archived_at"]; ok && archivedAtStr != "" {
		if t, err := parseTime(archivedAtStr); err == nil && t != nil {
			material.ArchivedAt = t
		}
	}
	if deletedAtStr, ok := data["deleted_at"]; ok && deletedAtStr != "" {
		if t, err := parseTime(deletedAtStr); err == nil && t != nil {
			material.DeletedAt = t
		}
	}

	return material, nil
}

func parseInt32(s string) int32 {
	var i int32
	_, err := fmt.Sscanf(s, "%d", &i)
	if err != nil {
		return 0
	}
	return i
}
