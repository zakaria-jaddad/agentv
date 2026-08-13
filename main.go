package main

import (
	"flag"
	"fmt"
	"os"
)

var token string

// ./agentv --token=*******
func main() {

	flag.StringVar(&token, "token", "", "Authentication token")
	flag.Parse()

	if token == "" {
		fmt.Fprintln(os.Stderr, "Usage: agentv <options>")
		return
	}
	fmt.Println("token: ", token)
}
