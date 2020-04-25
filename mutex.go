package redis_mutex

import (
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
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

var (
	// ErrLockFailed means failed to acquire the named lock.
	ErrLockFailed = errors.New("already locked")

	// ErrUnlockFailed means client not able to unlock because the lock is not owned by it.
	ErrUnlockFailed = errors.New("no lock")
)

// Mutex is a mutex based on redis.
type Mutex struct {
	id             string
	name           string
	acquireTimeout time.Duration
	lifetime       time.Duration
	pool           *redis.Pool
}

// Option is an option for tweaking Mutex.
type Option interface {
	Apply(*Mutex)
}

// OptionFunc is a function that configures a Mutex.
type OptionFunc func(*Mutex)

// Apply implements Option interface.
func (f OptionFunc) Apply(m *Mutex) { f(m) }

// WithAcquireTimeout returns an option that configures Mutex.acquireTimeout.
func WithAcquireTimeout(timeout time.Duration) Option {
	return OptionFunc(func(m *Mutex) {
		m.acquireTimeout = timeout
	})
}

// WithLockLifetime returns an option that configures Mutex.lifetime.
func WithLockLifetime(lifetime time.Duration) Option {
	return OptionFunc(func(m *Mutex) {
		m.lifetime = lifetime
	})
}

// NewMutex creates a new Mutex instance.
func NewMutex(redisPool *redis.Pool, name string, opts ...Option) *Mutex {
	m := &Mutex{
		id:             strconv.FormatInt(time.Now().UnixNano(), 10),
		name:           "mutex_" + name,
		acquireTimeout: time.Duration(0),
		pool:           redisPool,
	}

	for _, opt := range opts {
		opt.Apply(m)
	}
	return m
}

// String implements print interface.
func (m *Mutex) String() string {
	return m.name + "#" + m.id
}

// Lock locks the named resource.
func (m *Mutex) Lock() error {
	conn := m.pool.Get()
	defer conn.Close()

	// Set "mutex_xxx" key (the lock name).
	// PX: Set the specified expire time, in milliseconds.
	// NX: Only set the key if it does not already exist.
	reply, err := conn.Do("SET", m.name, m.id, "PX", m.lifetime.Nanoseconds()/int64(time.Millisecond), "NX")
	if err != nil {
		return errors.WithMessage(err, "redis SET lock")
	}
	if v, ok := reply.(string); ok && v == "OK" {
		return nil
	}

	return ErrLockFailed
}

// UnLock unlocks the named resource.
// Attempts to remove the key so long as the value matches.
// When the key not found or the value was incorrect, unlock will fail.
func (m *Mutex) UnLock() error {
	conn := m.pool.Get()
	defer conn.Close()

	command := redis.NewScript(1, unlockScript)
	ret, err := redis.Int(command.Do(conn, m.name, m.id))
	if err != nil {
		return errors.WithMessage(err, "redis do unlock script")
	}
	if ret != 1 {
		return ErrUnlockFailed
	}

	return nil
}
