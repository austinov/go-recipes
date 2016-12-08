package redigo

import (
	"log"

	"github.com/austinov/go-recipes/slack-bot/common"
	"github.com/austinov/go-recipes/slack-bot/config"
	"github.com/austinov/go-recipes/slack-bot/store"
	"github.com/garyburd/redigo/redis"
)

type Dao struct {
	pool *redis.Pool
}

func New(cfg config.DBConfig) store.Dao {
	pool := common.GetRedisPool(cfg.Network, cfg.Address, cfg.Password)
	return &Dao{
		pool,
	}
}

func (d *Dao) Close() error {
	log.Println("Close Dao in redigo package")
	d.pool.Close()
	return nil
}

func (d *Dao) GetCalendar(band string, from, to int64) ([]store.Event, error) {
	// TODO
	return nil, nil
}
