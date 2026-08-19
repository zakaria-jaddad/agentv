package backend

import (
	"fmt"
	"log"

	"github.com/zishang520/socket.io/clients/socket/v3"
)

func (c *Client) registerSocketEvents() {
	c.socket.On("agent:authenticated", func(args ...any) {
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
}

func (c *Client) Connect(socketURL string, agentID int64) error {

	opts := socket.DefaultOptions()
	opts.SetAuth(map[string]any{
		"token":   c.installationToken,
		"agentID": agentID,
	})
	s, err := socket.Connect(socketURL, opts)
	if err != nil {
		return fmt.Errorf("error connecting to backend: %w", err)
	}

	c.socket = s

	c.registerSocketEvents()

	return nil
}

func (c *Client) Emit(event string, args ...any) error {
	if c.socket == nil {
		return fmt.Errorf("backend socket is not connected")
	}

	return c.socket.Emit(event, args)
}
