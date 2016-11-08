package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"
	"sample/cui/client/handler"
	"sample/cui/client/handler/net"
	_ "sample/cui/client/view/term"
	"strings"
	"time"
)

// TODO network from flags
// TODO address from flags

func main() {
	// Ask for nick name as a peer name
	peerName := getUserName()

	ctrl := net.New("tcp", ":8822")
	//view := term.New(peerName, ctrl)
	view := &simpleView{
		s: ctrl,
	}

	if err := ctrl.Init(view); err != nil {
		log.Fatal(err)
	}
	defer ctrl.Disconnect()

	vch := view.Show()

	if len(os.Args) < 2 {
		if err := ctrl.BookRoom(peerName); err != nil {
			log.Fatal(errors.New(fmt.Sprintf("Error booking room: %s", err)))
		}
	} else {
		if err := ctrl.JoinRoom(os.Args[1], peerName); err != nil {
			log.Fatal(errors.New(fmt.Sprintf("Error joining room: %s", err)))
		}
	}

	<-vch
}

type simpleView struct {
	s handler.Sender
}

func (v *simpleView) ReceiveMessage(from, message string) {
	log.Println("ReceiveMessage", from, message)
}

func (v *simpleView) UpdatePeers(p []string) {
	log.Println("UpdatePeers", p)
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

//TODO extract
func getUserName() string {
	u, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(os.Stdin)
	name := ""
	for name == "" {
		fmt.Printf("Please type your nick name [%s]: ", u.Username)
		name, _ = reader.ReadString('\n')
		if len(name) > 0 {
			name = strings.TrimRight(name, "\n")
		}
		if name == "" {
			name = u.Username
		}
	}
	return name
}
