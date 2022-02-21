// distlock provides simple distributed locks using redis, mysql, postgresql,
// mongodb, etc.
package distlock

import (
	"strconv"
	"time"
)

type mutex struct {
	ns       string // namespace, as prefix for the lock name
	id       string
	owner    string
	lifetime time.Duration

	provider Provider
}

func (m *mutex) lockId() string {
	return m.ns + ":" + m.id
}

func (m *mutex) GetId() string {
	return m.lockId()
}

func (m *mutex) GetOwner() string {
	return m.owner
}

func (m *mutex) GetLifetime() time.Duration {
	return m.lifetime
}

// Option is an option for tweaking Mutex.
type Option interface {
	Apply(*mutex)
}

type optionFunc func(*mutex)

// Apply implements Option interface.
func (f optionFunc) Apply(m *mutex) { f(m) }

// WithNamespace returns an option that configures Mutex.ns.
func WithNamespace(ns string) Option {
	return optionFunc(func(m *mutex) {
		m.ns = ns
	})
}

// WithLockLifetime returns an option that configures Mutex.lifetime.
func WithLockLifetime(lifetime time.Duration) Option {
	return optionFunc(func(m *mutex) {
		m.lifetime = lifetime
	})
}

// NewMutex creates a new Mutex instance.
func newMutex(provider Provider, id string, opts ...Option) Mutex {
	m := &mutex{
		id:       id,
		owner:    strconv.FormatInt(time.Now().UnixNano(), 10),
		provider: provider,
	}

	WithNamespace("default").Apply(m)
	for _, opt := range opts {
		opt.Apply(m)
	}
	return m
}

// String implements print interface.
func (m *mutex) String() string {
	return "Mutex(" + m.provider.Name() + ":" + m.GetId() + ")"
}

// Lock locks the named resourc
func (m *mutex) Lock() error {
	return m.provider.Lock(m)
}

// Unlock unlocks the named resource.
// Attempts to remove the key so long as the value matches.
// When the key not found or the value was incorrect, unlock will fail.
func (m *mutex) Unlock() error {
	return m.provider.Unlock(m)
}
