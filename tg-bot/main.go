package main

import (
	"flag"
	"log"
	"time"

	"github.com/austinov/go-recipes/tg-bot/telegram"
)

func main() {
	var token string
	flag.StringVar(&token, "t", "", "telegram token")
	flag.Parse()

	b := telegram.New(token)
	go func() {
		<-time.After(1 * time.Minute)
		b.Stop()
	}()
	b.Start()
	log.Println("Stop the bot")

	log.Println("Start again")
	go func() {
		<-time.After(1 * time.Minute)
		b.Stop()
	}()
	b.Start()
	log.Println("Stop final")
}
