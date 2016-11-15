package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"strings"

	"github.com/austinov/go-recipes/termo-chat/client/handler/net"
	_ "github.com/austinov/go-recipes/termo-chat/client/view/simple"
	"github.com/austinov/go-recipes/termo-chat/client/view/term"
)

func main() {
	var (
		netw  = "tcp"
		laddr string
		room  string
	)
	flag.StringVar(&room, "room", "",
		"Room identity")

	flag.StringVar(&laddr, "addr", ":8822",
		"The syntax of addr is \"host:port\", like \"127.0.0.1:8822\". "+
			"If host is omitted, as in \":8822\", Listen listens on all available interfaces.")

	flag.Parse()

	un := getUserName()

	hdl := net.New(netw, laddr, un)
	view := term.New(un, hdl)
	//view := simple.New(hdl)

	hdl.Init(view, room)
}

func getUserName() string {
	u, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Please type your user name [%s]: ", u.Username)
	name, _ := reader.ReadString('\n')
	if len(name) > 0 {
		name = strings.TrimRight(name, "\n")
	}
	if name == "" {
		name = u.Username
	}
	// clear screen
	fmt.Print("\033[H\033[2J")
	return name
}
