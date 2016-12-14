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

func (d *Dao) GetCityEvents(city string, from, to, offset, limit int64) ([]store.Event, error) {
	// TODO
	return nil, nil
}

func (d *Dao) GetBandEvents(band string, from, to, offset, limit int64) ([]store.Event, error) {
	// TODO
	return nil, nil
}

func (d *Dao) GetBandInCityEvents(band string, city string, from, to, offset, limit int64) ([]store.Event, error) {
	// TODO
	return nil, nil
}
