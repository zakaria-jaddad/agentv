package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/zakaria-jaddad/agentv/internal/agentv"
	"github.com/zakaria-jaddad/agentv/internal/backend"
	"github.com/zakaria-jaddad/agentv/internal/bridge"
	"github.com/zakaria-jaddad/agentv/internal/config"
	"github.com/zakaria-jaddad/agentv/internal/manager"
)

func main() {

	// NOTE: ifnore for now
	os.MkdirAll("/tmp/agentv", 0750)
	// ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	ctx := context.Background()

	var confpath string
	flag.StringVar(&confpath, "config", "/etc/agentv/agentv.yml", "Configuration file path")
	flag.Parse()

	// Load agentv configuration
	conf, err := config.Load(confpath)
	if err != nil {
		log.Fatal(err)
	}

	// Validate Configuration
	if err := conf.Validate(); err != nil {
		log.Fatal(err)
	}

	// Creating Agent Object
	agentv := agentv.New(conf.Agent.Name)
	err = agentv.DiscoverSystemInfo()
	if err != nil {
		log.Fatal(err)
	}

	// Manager Creation
	// manager := manager.New(agentv, conf)

	// Agent Authentication
	log.Printf("Agent starting: %s on %s (%s/%s)",
		agentv.Name, agentv.Hostname, agentv.OS, agentv.Architecture)
	backendClient := backend.NewClient(conf.Backend.URL, conf.Token)
	log.Printf("Authenticating with backend: %s", conf.Backend.URL)

	auth, err := backendClient.AuthenticateWithRetry(agentv, ctx)
	if err != nil {
		log.Fatalf("%v", err)
	}
	log.Printf("Agent Authenticated as %d\n", auth.Data.AgentID)

	agentv.ID = auth.Data.AgentID

	// Socket io Connection
	log.Printf("Connecting To Backend Via WebSocket: %s\n", auth.Data.SocketURL)
	if err := backendClient.Connect(auth.Data.SocketURL, auth.Data.AgentID); err != nil {
		log.Fatalf("%v", err)
	}
	log.Printf("Successfully Connected To Backend Cia WebSocket: %s\n", auth.Data.SocketURL)

	// Unix Socket Connection
	log.Printf("Creating A Unix Socket: %s\n", conf.Vector.Socket)
	bridgeServer := bridge.NewServer(conf.Vector.Socket)

	log.Printf("Starting A Unix Socket: %s\n", conf.Vector.Socket)
	if err := bridgeServer.Start(ctx); err != nil {
		log.Fatalf("%v", err)
	}
	log.Printf("Unix Socket Listening On: %s\n", conf.Vector.Socket)

	// Creating The manager
	manager := manager.New(agentv, conf, backendClient, bridgeServer)

	select {}
	// Next:
	// 2. Start Vector
	// 3. Start heartbeat
}
