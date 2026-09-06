package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/zakaria-jaddad/agentv/internal/agentv"
	"github.com/zakaria-jaddad/agentv/internal/backend"
	"github.com/zakaria-jaddad/agentv/internal/bridge"
	"github.com/zakaria-jaddad/agentv/internal/config"
	"github.com/zakaria-jaddad/agentv/internal/manager"
)

// how am it supposed to run multiple instances of the same code with different information
// the same code but different env variables
func main() {
	// TODO: Check if the dir already exist
	os.MkdirAll("/tmp/agentv", 0750)

	log.Println("Process id: ", os.Getegid())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	log.Printf("Connecting To Backend Via WebSocket: %s\n", auth.Data.SocketURL)
	if err := backendClient.Connect(auth.Data.SocketURL, auth.Data.AgentID); err != nil {
		log.Fatalf("%v", err)
	}
	log.Printf("Successfully Connected To Backend Cia WebSocket: %s\n", auth.Data.SocketURL)

	log.Printf("Creating A Unix Socket: %s\n", conf.Vector.Socket)
	bridgeServer := bridge.NewServer(conf.Vector.Socket)

	log.Printf("Starting A Unix Socket: %s\n", conf.Vector.Socket)
	if err := bridgeServer.Start(ctx); err != nil {
		log.Fatalf("%v", err)
	}
	log.Printf("Unix Socket Listening On: %s\n", conf.Vector.Socket)

	manager := manager.New(agentv, conf, backendClient, bridgeServer)

	manager.RegisterEvents()

	// Start manager (Vector runs in background automatically)
	if err := manager.Start(ctx); err != nil {
		log.Fatalf("Failed to start manager: %v", err)
	}

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	log.Println("Shutting down...")
	cancel()
	if err := manager.Stop(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}
