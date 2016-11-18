package main

import (
	"flag"

	"github.com/austinov/go-recipes/tg-bot/telegram"
)

func main() {
	var token string
	flag.StringVar(&token, "t", "", "telegram token")
	flag.Parse()

	b := telegram.New(token)
	b.Start()
}
