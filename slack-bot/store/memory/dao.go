package memory

import (
	"log"

	"github.com/austinov/go-recipes/slack-bot/config"
	"github.com/austinov/go-recipes/slack-bot/store"
)

type Dao struct {
}

func New(cfg config.DBConfig) store.Dao {
	return &Dao{}
}

func (d *Dao) Close() error {
	log.Println("Close Dao in memory package")
	return nil
}

func (d *Dao) GetCalendar(band string, from, to int64) ([]store.Event, error) {
	// TODO
	return nil, nil
}
