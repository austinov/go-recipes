package memory

import (
	"github.com/austinov/go-recipes/slack-bot/config"
	"github.com/austinov/go-recipes/slack-bot/store"
)

type Dao struct {
}

func New(cfg config.DBConfig) store.Dao {
	return &Dao{}
}

func (d *Dao) Close() error {
	return nil
}

func (d *Dao) AddBandEvents(events []store.Event) error {
	// TODO
	return nil
}

func (d *Dao) GetEvents(band string, city string, from, to int64, offset, limit int) ([]store.Event, error) {
	// TODO
	return nil, nil
}
