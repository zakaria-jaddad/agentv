package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/zakaria-jaddad/agentv/internal/backend"
	"github.com/zakaria-jaddad/agentv/internal/manager"
)

func main() {

	var confpath string
	flag.StringVar(&confpath, "config", "/etc/agentv/agentv.yml", "Configuration file path")
	flag.Parse()

	manager, err := manager.New(confpath)
	if err != nil {
		log.Fatalf("%v", err)
	}

	log.Printf("Agent starting: %s on %s (%s/%s)",
		manager.Agent.Name, manager.Agent.Hostname, manager.Agent.OS, manager.Agent.Architecture)
	client := backend.NewClient(manager.Config.Backend.URL, manager.Config.Token)

	log.Printf("Authenticating with backend: %s", manager.Config.Backend.URL)
	ctx := context.Background()
	auth, err := client.AuthenticateWithRetry(manager, ctx)
	if err != nil {
		log.Fatalf("%v", err)
	}
	log.Printf("Agent Authenticated as %d\n", auth.Data.AgentID)

	manager.Agent.ID = auth.Data.AgentID

	fmt.Println(auth.Data.SocketURL)

	client.Connect(auth.Data.SocketURL, auth.Data.AgentID)

	client.Emit("agent:hello", "hello")

	select {}
	// Next:
	// 1. Connect Socket.IO using auth.SessionToken
	// 2. Start Vector
	// 3. Start heartbeat
}
