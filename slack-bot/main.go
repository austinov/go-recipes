package main

import (
	"flag"
	"github.com/austinov/go-recipes/slack-bot/bot"
)

func main() {
	var token string
	flag.StringVar(&token, "t", "", "slack token")
	flag.Parse()

	b := bot.New(token)
	b.Start()
}
