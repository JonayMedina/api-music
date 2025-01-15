package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(uri string) (*RedisClient, error) {
	opt, err := redis.ParseURL(uri)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	// Verificar conexión
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisClient{
		client: client,
	}, nil
}

func (rc *RedisClient) Close() error {
	return rc.client.Close()
}

// Get obtiene un valor del caché
func (rc *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := rc.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

// Set guarda un valor en el caché
func (rc *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return rc.client.Set(ctx, key, data, expiration).Err()
}

// Del elimina una o más claves del caché
func (rc *RedisClient) Del(ctx context.Context, keys ...string) error {
	return rc.client.Del(ctx, keys...).Err()
}
