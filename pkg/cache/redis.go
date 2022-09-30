//redis封装
package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/extra/redisotel/v8"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/attribute"
	"time"
)

type Option func(*option)

type option struct {
}

func newOption() *option {
	return &option{}
}

//封装缓存接口
type Repo interface {
	i()
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
	Expire(ctx context.Context, key string, ttl time.Duration) bool
	ExpireAt(ctx context.Context, key string, ttl time.Time) bool
	Del(ctx context.Context, key string) bool
	Exists(ctx context.Context, keys ...string) bool
	Incr(ctx context.Context, key string) int64
	Client() *redis.Client
	Close() error
}

type cacheRepo struct {
	serverName string
	client     *redis.Client
}

type RedisConf struct {
	Addr         string
	Pass         string
	Db           int
	MaxRetries   int
	PoolSize     int
	MinIdleConns int
}

func New(serverName string, cfg *RedisConf) (Repo, error) {
	client, err := redisConnect(serverName, cfg)
	if err != nil {
		return nil, err
	}
	return &cacheRepo{
		serverName: serverName,
		client:     client,
	}, nil
}

func redisConnect(serverName string, cfg *RedisConf) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Pass,
		DB:           cfg.Db,
		MaxRetries:   cfg.MaxRetries,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	client.AddHook(redisotel.NewTracingHook(redisotel.WithAttributes(
		attribute.String("servername", serverName),
	)))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, errors.New("Ping redis失败！")
	}

	return client, nil
}

func (c cacheRepo) i() {
}

func (c cacheRepo) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	var err error
	if err = c.client.Set(ctx, key, value, ttl).Err(); err != nil {
		err = errors.New(fmt.Sprintf("存入redis:%s失败！", key))
	}
	return err
}

func (c cacheRepo) Get(ctx context.Context, key string) (string, error) {
	var err error
	value, err := c.client.Get(ctx, key).Result()
	if err != nil {
		err = err
	}
	return value, err
}

func (c cacheRepo) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return -1, errors.New(fmt.Sprintf("获取redis:%s失败！", key))
	}
	return ttl, nil
}

func (c cacheRepo) Expire(ctx context.Context, key string, ttl time.Duration) bool {
	ok, _ := c.client.Expire(ctx, key, ttl).Result()
	return ok
}

func (c cacheRepo) ExpireAt(ctx context.Context, key string, ttl time.Time) bool {
	ok, _ := c.client.ExpireAt(ctx, key, ttl).Result()
	return ok
}

func (c cacheRepo) Del(ctx context.Context, key string) bool {
	var err error
	if key == "" {
		return true
	}

	value, err := c.client.Del(ctx, key).Result()
	if err != nil {
		err = errors.New(fmt.Sprintf("删除失败:%s!", key))
	}
	return value > 0
}

func (c cacheRepo) Exists(ctx context.Context, keys ...string) bool {
	if len(keys) == 0 {
		return true
	}

	value, _ := c.client.Exists(ctx, keys...).Result()
	return value > 0
}

func (c cacheRepo) Incr(ctx context.Context, key string) int64 {
	var err error
	value, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		err = errors.New(fmt.Sprintf("incr key:%s,失败!", key))
	}
	return value
}

func (c cacheRepo) Client() *redis.Client {
	return c.client
}

func (c cacheRepo) Close() error {
	return c.client.Close()
}
