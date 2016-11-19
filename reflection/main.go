package main

import (
	"fmt"

	"github.com/austinov/go-recipes/reflection/trick"
)

func main() {
	msg := trick.Message{
		Response: trick.Response{
			Code: 200,
			Desc: "OK",
		},
		Text: "Hello!",
	}
	resp, ok := trick.GetEmbedded(msg)
	fmt.Printf("\"Message has Response\" is %v: %#v\n", ok, resp)
	people := trick.People{
		Name:     "Yuri Alekseyevich Gagarin",
		Birthday: "9 March 1934",
	}
	resp, ok = trick.GetEmbedded(people)
	fmt.Printf("\"People has Response\" is %v: %#v\n", ok, resp)
}
