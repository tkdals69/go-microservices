package cache

import (
    "context"
    "github.com/redis/go-redis/v9"
    "time"
)

type RedisCache struct {
    client *redis.Client
    ctx    context.Context
}

func NewRedisCache(addr string, password string, db int) *RedisCache {
    rdb := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       db,
    })

    return &RedisCache{
        client: rdb,
        ctx:    context.Background(),
    }
}

func (c *RedisCache) Set(key string, value interface{}, expiration time.Duration) error {
    return c.client.Set(c.ctx, key, value, expiration).Err()
}

func (c *RedisCache) Get(key string) (string, error) {
    return c.client.Get(c.ctx, key).Result()
}

func (c *RedisCache) Delete(key string) error {
    return c.client.Del(c.ctx, key).Err()
}

func (c *RedisCache) Close() error {
    return c.client.Close()
}