package main

import (
	"github.com/austinov/go-recipes/slack-bot/bot"
	"github.com/austinov/go-recipes/slack-bot/config"
	"github.com/austinov/go-recipes/slack-bot/dao/redigo"
)

func main() {
	cfg := config.GetConfig()

	dao := redigo.New(cfg.Db)
	defer dao.Close()

	b := bot.New(cfg.Bot.Token, dao)
	b.Start()
}
