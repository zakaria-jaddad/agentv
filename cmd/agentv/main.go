package main

import (
	"flag"
	"log"

	"github.com/zakaria-jaddad/agentv/internal/backend"
	"github.com/zakaria-jaddad/agentv/internal/manager"
)

func main() {

	var confpath string
	flag.StringVar(&confpath, "config", "/etc/agentv/agentv.yml", "Configuration file path")
	flag.Parse()

	manager := manager.New(confpath)
	client := backend.NewClient(manager.Config.Backend.URL, manager.Config.Token)

	log.Printf("Authenticating agent\n")
	auth, err := client.Authenticate(manager)
	if err != nil {
		log.Fatalf("%v", err)
	}
	log.Printf("Agent Authenticated as %d\n", auth.Data.AgentID)

	// Next:
	// 1. Connect Socket.IO using auth.SessionToken
	// 2. Start Vector
	// 3. Start heartbeat
}
