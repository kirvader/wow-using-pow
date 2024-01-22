package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var _ StorageSet = &RedisClient{}

// Here I implemented a quick middleware to access only those parts of redis which I need
type RedisClient struct {
	client *redis.Client
}

const (
	DefaultRedisValue = "present"
)

func NewRedisCache(ctx context.Context, address string) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: address,
	})

	status := rdb.Ping(ctx)
	if status.Err() != nil {
		return nil, fmt.Errorf("couldn't ping redis cache: %v", status.Err())
	}

	return &RedisClient{
		client: rdb,
	}, nil
}

func (rc *RedisClient) Add(ctx context.Context, key string, duration time.Duration) error {
	return rc.client.Set(ctx, key, DefaultRedisValue, duration).Err()
}

func (rc *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	res, err := rc.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return res > 0, nil
}

func (rc *RedisClient) Delete(ctx context.Context, key string) error {
	return rc.client.Del(ctx, key).Err()
}

func (rc *RedisClient) Close(ctx context.Context) error {
	return rc.client.Close()
}
