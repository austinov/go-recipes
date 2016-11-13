package simple

import (
	"github.com/austinov/go-recipes/termo-chat/client/handler"
	"github.com/austinov/go-recipes/termo-chat/client/view"
	"log"
	"time"
)

type simpleView struct {
	s handler.Sender
}

func New(sender handler.Sender) view.View {
	return &simpleView{
		s: sender,
	}
}

func (v *simpleView) ViewRoom(id string) {
	log.Println("ViewRoom", id)
}

func (v *simpleView) ViewMessage(kind byte, from, msg string) {
	log.Println("ViewMessage", kind, from, msg)
}

func (v *simpleView) UpdatePeers(peers []string) {
	log.Println("UpdatePeers", peers)
}
func (v *simpleView) Show() <-chan struct{} {
	quit := make(chan struct{})
	go func() {
		<-time.After(10 * time.Second)
		v.s.SendMessage("Hello!")
		<-time.After(30 * time.Second)
		quit <- struct{}{}
	}()
	return quit
}

func (v *simpleView) Quit() {
	log.Println("Quit")
}
