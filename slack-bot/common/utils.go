package common

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

const oneDay = (24*60*60 - 1) * time.Second

// BeginOfDate returns begin of date
func BeginOfDate(d time.Time) time.Time {
	t := d.Truncate(time.Hour)
	return t.Add(time.Duration(-1*t.Hour()) * time.Hour)
}

// EndOfDate returns end of date - begin date minus one second
func EndOfDate(d time.Time) time.Time {
	t := d.Truncate(time.Hour)
	return t.Add(time.Duration(-1*t.Hour()) * time.Hour).Add(oneDay)
}

// GetRedisPool returns connection pool for redis
func GetRedisPool(network, address, password string) (pool *redis.Pool) {
	pool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(network, address)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, _ time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	return pool
}
