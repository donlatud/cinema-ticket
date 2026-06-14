package booking

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/cinema-booking/backend/internal/lock"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	return mr, client
}

func TestConcurrentLock(t *testing.T) {
	_, client := setupTestRedis(t)
	locker := lock.NewRedisLock(client, 300*time.Second)

	ctx := context.Background()
	showtimeID := "507f1f77bcf86cd799439011"
	seatNo := "A5"

	var wg sync.WaitGroup
	var success atomic.Int32

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := locker.AcquireLock(ctx, showtimeID, seatNo, uuid.New().String())
			if err == nil {
				success.Add(1)
			}
		}()
	}

	wg.Wait()

	if got := success.Load(); got != 1 {
		t.Fatalf("expected exactly 1 successful lock, got %d", got)
	}
}

func TestLockExpiresAndCanBeReacquired(t *testing.T) {
	mr, client := setupTestRedis(t)
	locker := lock.NewRedisLock(client, 10*time.Second)

	ctx := context.Background()
	showtimeID := "507f1f77bcf86cd799439011"
	seatNo := "A5"

	firstToken, err := locker.AcquireLock(ctx, showtimeID, seatNo, "user-1")
	if err != nil {
		t.Fatalf("first acquire: %v", err)
	}

	_, err = locker.AcquireLock(ctx, showtimeID, seatNo, "user-2")
	if !errors.Is(err, lock.ErrLockNotAcquired) {
		t.Fatalf("expected ErrLockNotAcquired while locked, got %v", err)
	}

	mr.FastForward(11 * time.Second)

	secondToken, err := locker.AcquireLock(ctx, showtimeID, seatNo, "user-2")
	if err != nil {
		t.Fatalf("re-acquire after TTL: %v", err)
	}
	if secondToken == firstToken {
		t.Fatal("expected a new lock token after TTL expiry")
	}

	_, err = locker.AcquireLock(ctx, showtimeID, seatNo, "user-3")
	if !errors.Is(err, lock.ErrLockNotAcquired) {
		t.Fatalf("expected seat to stay locked after re-acquire, got %v", err)
	}
}

func TestReleaseLockRequiresMatchingToken(t *testing.T) {
	_, client := setupTestRedis(t)
	locker := lock.NewRedisLock(client, 300*time.Second)

	ctx := context.Background()
	showtimeID := "507f1f77bcf86cd799439011"
	seatNo := "A5"

	ownerToken, err := locker.AcquireLock(ctx, showtimeID, seatNo, "user-1")
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}

	if err := locker.ReleaseLock(ctx, showtimeID, seatNo, "wrong-token"); err != nil {
		t.Fatalf("release with wrong token: %v", err)
	}

	_, err = locker.AcquireLock(ctx, showtimeID, seatNo, "user-2")
	if !errors.Is(err, lock.ErrLockNotAcquired) {
		t.Fatalf("expected lock to remain after wrong-token release, got %v", err)
	}

	if err := locker.ReleaseLock(ctx, showtimeID, seatNo, ownerToken); err != nil {
		t.Fatalf("release with owner token: %v", err)
	}

	_, err = locker.AcquireLock(ctx, showtimeID, seatNo, "user-2")
	if err != nil {
		t.Fatalf("expected lock to be available after owner release, got %v", err)
	}
}
