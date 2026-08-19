package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/zakaria-jaddad/agentv/internal/backend"
	"github.com/zakaria-jaddad/agentv/internal/manager"
	"github.com/zishang520/socket.io/clients/socket/v3"
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
	client := backend.NewClient(manager.Config.Backend.AuthEndpoint, manager.Config.Token)

	log.Printf("Authenticating with backend: %s", manager.Config.Backend.AuthEndpoint)
	auth, err := client.Authenticate(manager)
	if err != nil {
		log.Fatalf("%v", err)
	}
	log.Printf("Agent Authenticated as %d\n", auth.Data.AgentID)

	manager.Agent.ID = auth.Data.AgentID

	fmt.Println(auth.Data.SocketURL)

	opts := socket.DefaultOptions()
	opts.SetAuth(map[string]any{
		"token":   manager.Config.Token,
		"agentID": manager.Agent.ID,
	})
	socketClient, err := socket.Connect(auth.Data.SocketURL, opts)
	if err != nil {
		log.Fatalf("%v", err)
	}

	socketClient.On("agent:authenticated", func(args ...any) {
		if len(args) == 0 {
			log.Println("authenticated, but no payload received")
			return
		}

		if payload, ok := args[0].(map[string]any); ok {
			log.Printf("Connected to backend At=%v", payload["connectedAt"])
		} else {
			log.Printf("authenticated: %v", args[0])
		}
	})

	socketClient.Emit("agent:hello", "hello")

	select {}
	// Next:
	// 1. Connect Socket.IO using auth.SessionToken
	// 2. Start Vector
	// 3. Start heartbeat
}
