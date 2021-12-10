package distlock

import "time"

type NamedLock interface {
	GetId() string
	GetOwner() string
	GetLifetime() time.Duration
}

type Provider interface {
	Name() string
	Lock(info NamedLock) error
	Unlock(info NamedLock) error
}

type Mutex interface {
	Lock() error
	Unlock() error
	String() string
}

type MutexFactory interface {
	New(id string, opts ...Option) Mutex
}

type factory struct {
	provider    Provider
	defaultOpts []Option
}

func New(provider Provider, defaultOpts ...Option) MutexFactory {
	return &factory{
		provider:    provider,
		defaultOpts: defaultOpts,
	}
}

func (f *factory) New(id string, opts ...Option) Mutex {
	return newMutex(f.provider, id, append(f.defaultOpts, opts...)...)
}
