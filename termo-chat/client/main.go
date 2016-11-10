package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/user"
	"strings"

	"github.com/austinov/go-recipes/termo-chat/client/handler/net"
	"github.com/austinov/go-recipes/termo-chat/client/view/term"
)

// TODO network from flags
// TODO address from flags
// TODO preambula
// TODO error, info messages

func main() {
	userName := getUserName()
	room := ""
	if len(os.Args) > 1 {
		room = os.Args[1]
	}

	hdl := net.New("tcp", ":8822", userName)
	view := term.New(userName, hdl)

	if err := hdl.Init(view, room); err != nil {
		log.Fatal(err)
	}
}

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
