# distlock

**Distributed Locks implementation in Go**.

## Features

1. Namespace (names in the same namespace are unique, default namespace is `"default"`)
2. Auto/No expiration (auto-released after a specific time or never expire)
3. Can work with multiple backends:
   - [x] Redis
   - [x] MySQL
   - [ ] Postgres

## Usage

```go
import (
    "github.com/ggicci/distlock"
    "github.com/gomodule/redigo/redis"
)

var redisPool = &redis.Pool{
    // ... configure your redis client
}

var heavyRequestsGuard = distlock.New(
    distlock.NewRedisProvider(redisPool),
    distlock.WithNamespace("heavy_requests"),
    distlock.WithLockLifetime(10 * time.Second), // lifetime: 10s
)

user := session.CurrentUser()
mu := heavyRequestsGuard.New(
    user.Username,
    WithLockLifetime(time.Minute), // override the default lifetime option: 10s
)
if err := mu.Lock(); err != nil {
    return err
}
defer mu.Unlock()

// do sth.
```
