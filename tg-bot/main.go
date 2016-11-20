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
		<-time.After(10 * time.Second)
		b.Stop()
	}()
	b.Start()
	log.Println("Try new start")
	go func() {
		<-time.After(10 * time.Second)
		b.Stop()
	}()
	b.Start()
	log.Println("Stop all")
}
