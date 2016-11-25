package main

import (
	"flag"
	"log"
	"time"

	"github.com/austinov/go-recipes/tg-bot/bot"
)

func main() {
	var token string
	flag.StringVar(&token, "t", "", "telegram token")
	flag.Parse()

	b := bot.New(token)
	go func() {
		<-time.After(1 * time.Minute)
		log.Println("Stop telegram bot.")
		b.Stop()
	}()
	log.Println("Start telegram bot.")
	b.Start()

	log.Println("Start telegram bot again.")
	b.Start()
}
