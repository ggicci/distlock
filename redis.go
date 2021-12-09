package distlock

import (
	"errors"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

const (
	unlockScript = `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end
	`
)

// Pool represents a pool of redis connections.
type Pool interface {
	// Get returns a connection from the pool.
	Get() redis.Conn
}

type redisProvider struct {
	pool Pool
}

func NewRedisProvider(redisPool Pool) Provider {
	return &redisProvider{
		pool: redisPool,
	}
}

func (p *redisProvider) Name() string {
	return "redis"
}

func (p *redisProvider) Lock(lock LockInfo) error {
	conn := p.pool.Get()
	defer conn.Close()

	// SET key value PX milliseconds NX
	// PX: Set the specified expire time, in milliseconds.
	// NX: Only set the key if it does not already exist.
	reply, err := conn.Do(
		"SET", lock.GetLockId(), lock.GetLockOwner(),
		"PX", lock.GetLifetime().Nanoseconds()/int64(time.Millisecond),
		"NX",
	)
	if err != nil {
		return fmt.Errorf("redis SET: %w", err)
	}
	if v, ok := reply.(string); ok && v == "OK" {
		return nil
	}

	return errors.New("already locked")
}

func (p *redisProvider) Unlock(lock LockInfo) error {
	conn := p.pool.Get()
	defer conn.Close()

	command := redis.NewScript(1, unlockScript)
	ret, err := redis.Int(command.Do(conn, lock.GetLockId(), lock.GetLockOwner()))
	if err != nil {
		return fmt.Errorf("redis EVAL: %w", err)
	}
	if ret == 0 {
		return errors.New("lock not found")
	}
	return nil
}
