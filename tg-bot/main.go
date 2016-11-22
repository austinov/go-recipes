package main

import (
	"flag"
	"log"
	"time"

	"github.com/austinov/go-recipes/tg-bot/api"
)

func main() {
	var token string
	flag.StringVar(&token, "t", "", "telegram token")
	flag.Parse()

	b := api.New(token)
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
