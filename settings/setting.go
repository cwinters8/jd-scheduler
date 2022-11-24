package settings

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

type Setting struct {
	Key   string
	Value string
}

// get a new setting
func New(key string, value string) *Setting {
	return &Setting{
		Key:   key,
		Value: value,
	}
}

// insert or update a setting
func (s *Setting) Save(ctx context.Context, db *redis.Client) error {
	return db.Set(ctx, s.Key, s.Value, 0).Err()
}

// retrieve a setting
func Get(ctx context.Context, key string, db *redis.Client) (*Setting, error) {
	setting := Setting{Key: key}
	value, err := db.Get(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get value for key %q: %w", key, err)
	}
	setting.Value = value
	return &setting, nil
}
