package lock

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var ErrLockNotAcquired = errors.New("lock not acquired")

var releaseScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
else
	return 0
end
`)

type RedisLock struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisLock(client *redis.Client, ttl time.Duration) *RedisLock {
	return &RedisLock{client: client, ttl: ttl}
}

func lockKey(showtimeID, seatNo string) string {
	return fmt.Sprintf("lock:seat:%s:%s", showtimeID, seatNo)
}

func (l *RedisLock) AcquireLock(ctx context.Context, showtimeID, seatNo, userID string) (string, error) {
	_ = userID
	token := uuid.New().String()
	key := lockKey(showtimeID, seatNo)

	ok, err := l.client.SetNX(ctx, key, token, l.ttl).Result()
	if err != nil {
		return "", err
	}
	if !ok {
		return "", ErrLockNotAcquired
	}
	return token, nil
}

func (l *RedisLock) ReleaseLock(ctx context.Context, showtimeID, seatNo, token string) error {
	key := lockKey(showtimeID, seatNo)
	result, err := releaseScript.Run(ctx, l.client, []string{key}, token).Int64()
	if err != nil {
		return err
	}
	if result == 0 {
		return nil
	}
	return nil
}
