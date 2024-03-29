# distlock

[![Go](https://github.com/ggicci/distlock/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/ggicci/distlock/actions/workflows/go.yml) [![codecov](https://codecov.io/gh/ggicci/distlock/branch/main/graph/badge.svg?token=2MDBW1V2TI)](https://codecov.io/gh/ggicci/distlock) [![Go Reference](https://pkg.go.dev/badge/github.com/ggicci/distlock.svg)](https://pkg.go.dev/github.com/ggicci/distlock)

<div align="center">
   Simple <b>Distributed Locks</b> implementation in Go
   <div>backed on</div>
   <div>Redis, MySQL, PostgreSQL, MongoDB etc.</div>
</div>

## Features

1. Namespace (names in the same namespace are unique, default namespace is `"default"`)
2. Auto/No expiration (auto-released after a specific time or never expire)
3. Support multiple backends:
   - [x] Redis
   - [x] MySQL
   - [x] PostgreSQL
   - [ ] MongoDB

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
