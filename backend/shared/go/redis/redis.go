package redis_connector

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// key for basic user info: "basic_user_info:<id>" Returns models.User
// key for image url: "img_<variant>:<id>"

var (
	ErrNoConnection = errors.New("can't find redis")
	ErrFailedTest1  = errors.New("failed special connection test 1")
	ErrFailedTest2  = errors.New("failed special connection test 2")
	ErrFailedTest3  = errors.New("failed special connection test 3")
	ErrFailedTest4  = errors.New("failed special connection test 4")
	ErrNotFound     = errors.New("entry wasn't found")
	ErrIncrExpFail  = errors.New("incr func failed to add expiration, extremely unexpected!!!")
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(sentinelAddrs []string, password string, db int, masterName string) *RedisClient {

	redisClient := &RedisClient{
		client: redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    masterName,
			SentinelAddrs: sentinelAddrs,
			Password:      password,
			DB:            db,
		}),
	}
	// redisClient2 := &RedisClient{
	// 	client: redis.NewClient(&redis.Options{
	// 		Addr:     addr,
	// 		Password: password,
	// 		DB:       db,
	// 	}),
	// }

	return redisClient
}

// IncrEx increments the integer value of a key by one.
func (c *RedisClient) IncrEx(ctx context.Context, key string, expSeconds int64) (int, error) {
	//incrementing number on redis for key, creates it if it doesn't exist
	currentCount, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	//if it was created it will be at 1, so we add expiration date
	if currentCount == 1 {
		ok, err := c.client.Expire(ctx, key, time.Duration(expSeconds)*time.Second).Result()
		if err != nil {
			return 0, err
		}

		//somehow the key doesn't exist, even though we just created it...
		if !ok {
			return 0, ErrIncrExpFail
		}
	}
	return int(currentCount), err
}

// SetStr stores a string value under `key` with expiration `exp`.
func (c *RedisClient) SetStr(ctx context.Context, key string, value string, exp time.Duration) error {
	err := c.client.Set(ctx, key, value, exp).Err()
	return err
}

// GetStr retrieves a string value stored under `key`.
func (c *RedisClient) GetStr(ctx context.Context, key string) (any, error) {
	value, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, ErrNotFound
	}
	return value, err
}

// SetObj marshals a Go value to JSON and stores it under `key` with expiration `exp`.
func (c *RedisClient) SetObj(ctx context.Context, key string, value any, exp time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, b, exp).Err()
}

// GetObj retrieves the JSON stored at `key` and unmarshals it into `dest`.
// `dest` must be a pointer to the value to populate.
func (c *RedisClient) GetObj(ctx context.Context, key string, dest any) error {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return ErrNotFound
	}
	return json.Unmarshal([]byte(val), dest)
}

// Del deletes the specified key from Redis.
func (c *RedisClient) Del(ctx context.Context, key string) error {
	err := c.client.Del(ctx, key).Err()
	return err
}

// TestRedisConnection performs a series of operations to verify the Redis connection is functioning correctly.
func (c *RedisClient) TestRedisConnection() error {
	ctx := context.Background()
	ping, err := c.client.Ping(ctx).Result()
	if err != nil || ping != "PONG" {
		return errors.Join(ErrNoConnection, err)
	}

	err = c.SetStr(ctx, "test_key123", "value", time.Second)
	if err != nil {
		return errors.Join(ErrFailedTest1, err)
	}

	val, err := c.GetStr(ctx, "test_key123")
	valStr, ok := val.(string)
	if err != nil || !ok || valStr != "value" {
		return errors.Join(ErrFailedTest2, err)
	}

	err = c.Del(ctx, "test_key123")
	if err != nil {
		return errors.Join(ErrFailedTest3, err)
	}

	val, err = c.GetStr(ctx, "test_key123")
	if err != redis.Nil && val == "" {
		return errors.Join(ErrFailedTest4, err)
	}

	return nil
}
