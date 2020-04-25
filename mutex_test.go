package redis_mutex

import (
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
)

func newRedisConnection() (conn redis.Conn, err error) {
	return redis.Dial(
		"tcp",
		"127.0.0.1:6379",
		redis.DialPassword(""),
		redis.DialConnectTimeout(3*time.Second),
		redis.DialKeepAlive(time.Minute),
		redis.DialDatabase(0),
	)
}

func TestMutex(t *testing.T) {
	var (
		err error

		redisPool = &redis.Pool{
			MaxIdle: 10,
			Dial:    newRedisConnection,
		}

		m = NewMutex(redisPool, "test", WithLockLifetime(1*time.Second))
	)

	err = m.Lock()
	if err != nil {
		t.Errorf("Lock failed: name=%s, err=%v", m, err)
	}
	err = m.UnLock()
	if err != nil {
		t.Errorf("Unlock failed: name=%s, err=%v", m, err)
	}

	err = m.Lock()
	if err != nil {
		t.Errorf("Lock failed: name=%s, err=%v", m, err)
	}

	// Here unlock should fail because lock was released by lifetime.
	time.Sleep(1 * time.Second)
	err = m.UnLock()
	if err != ErrUnlockFailed {
		t.Errorf("unlock should fail")
	}

	m.Lock()
	// Here lock should fail.
	err = m.Lock()
	if err != ErrLockFailed {
		t.Errorf("lock should fail")
	}
	err = m.UnLock()
	if err != nil {
		t.Errorf("Unlock failed: name=%s, err=%v", m, err)
	}

	m.Lock()
	m.id = "2814"
	err = m.UnLock()
	if err != ErrUnlockFailed {
		t.Errorf("unlock should fail")
	}
}
