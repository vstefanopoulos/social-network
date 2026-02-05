package ratelimit

import (
	"context"
	"errors"
)

var ErrBadValue = errors.New("bad value received")
var ErrNotANumber = errors.New("value stored not a number")
var ErrStorageProblem = errors.New("problem with storage mechanism")

// what the rate limiter will be using to save its ratelimit entries
type storage interface {
	IncrEx(ctx context.Context, key string, durationSeconds int64) (currentCount int, err error)
}

type ratelimiter struct {
	globalPrefix string //will be used as a prefix on all keys, when the regular save function is called
	storage      storage
}

func NewRateLimiter(globalPrefix string, storage storage) *ratelimiter {
	rateLimiter := &ratelimiter{
		globalPrefix: globalPrefix,
		storage:      storage,
	}
	return rateLimiter
}

// Allow checks if the action identified by `key` is allowed under the rate limit defined by `limit` and `duration`. It will prefix the key with the global prefix.
func (rl *ratelimiter) Allow(ctx context.Context, key string, limit int, durationSeconds int64) (bool, error) {
	return rl.AllowRawKey(ctx, rl.globalPrefix+key, limit, durationSeconds)
}

// AllowRawKey checks if the action identified by `key` is allowed under the rate limit defined by `limit` and `duration`.
func (rl *ratelimiter) AllowRawKey(ctx context.Context, key string, limit int, durationSeconds int64) (bool, error) {
	storageLimit, err := rl.storage.IncrEx(ctx, key, durationSeconds)
	if err != nil {
		return false, errors.Join(ErrStorageProblem, err)
	}
	if storageLimit > limit {
		return false, nil
	}
	return true, nil
}
