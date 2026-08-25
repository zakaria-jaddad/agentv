package backend

import (
	"fmt"
	"log"

	"github.com/zakaria-jaddad/agentv/internal/bridge"
	"github.com/zishang520/socket.io/clients/socket/v3"
	"github.com/zishang520/socket.io/v3/pkg/types"
)

type EventHandler func() error

func (c *Client) Connect(socketURL string, agentID int) error {

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

	return nil
}

func (c *Client) RegisterEvent(event bridge.Event, eventHandler EventHandler) {

	c.socket.On(types.EventName(event), func(args ...any) {
		if eventHandler == nil {
			log.Printf("%s function not set, cannot handle %s", event, event)
			return
		}

		if err := eventHandler(); err != nil {
			log.Printf("failed to %s vector: %v", event, err)
		}
	})
}

func (c *Client) Emit(event string, args ...any) error {
	if c.socket == nil {
		return fmt.Errorf("backend socket is not connected")
	}

	return c.socket.Emit(event, args)
}

func (c *Client) Disconnect() {
	if c.socket != nil {
		c.socket.Disconnect()
	}
}
