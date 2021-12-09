package distlock

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

var (
	redisPool = &redis.Pool{
		MaxIdle: 10,
		Dial:    newRedisConnection,
	}
)

func TestRedisProvider(t *testing.T) {
	provider, _ := NewRedisProvider(redisPool)
	runBasicLockTests(t, provider)
}
