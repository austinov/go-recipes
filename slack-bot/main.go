package main

import (
	"log"

	"github.com/austinov/go-recipes/slack-bot/bot"
	"github.com/austinov/go-recipes/slack-bot/config"
	"github.com/austinov/go-recipes/slack-bot/loader/cmetal"
	"github.com/austinov/go-recipes/slack-bot/store"
	"github.com/austinov/go-recipes/slack-bot/store/memory"
	"github.com/austinov/go-recipes/slack-bot/store/redigo"
)

func main() {
	cfg := config.GetConfig()
	if err := cfg.Verify(); err != nil {
		log.Fatal(err)
	}

	dao := createDao(cfg.DB)
	defer dao.Close()

	l := cmetal.New(cfg.CMetal, dao)
	// start loader in separate go-routine
	go l.Start()

	b := bot.New(cfg.Bot, dao)
	// start bot and block until return
	b.Start()
	// stop loader
	l.Stop()
}

func createDao(cfg config.DBConfig) store.Dao {
	switch cfg.Type {
	case "redis":
		return redigo.New(cfg)
	case "memory":
		return memory.New(cfg)
	}
	log.Fatal("Unknown db type " + cfg.Type)
	return nil
}
