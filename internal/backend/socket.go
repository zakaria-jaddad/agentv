package backend

import (
	"errors"
	"fmt"
	"log"

	"github.com/zakaria-jaddad/agentv/internal/bridge"
	"github.com/zishang520/socket.io/clients/socket/v3"
	"github.com/zishang520/socket.io/v3/pkg/types"
)

var ErrBackendUnavailable = errors.New("backend socket is unavailable")

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
	socket := c.socket

	if socket == nil || !socket.Connected() || !isSocketWritable(socket) {
		return ErrBackendUnavailable
	}
	log.Printf("Emitting event: %s with args: %v", event, args)
	return c.socket.Emit(event, args...)
}

func (c *Client) Disconnect() {
	if c.socket != nil {
		c.socket.Disconnect()
	}
}

// isSocketWritable checks if the socket is writable.
// Socket io  has several layers
// isSocketWritable checks if the socket is writable by checking if the socket is connected and if the underlying transport is writable.
// A namespace socket can report that it is connected when the undrelying transport hat become unwritable or the connection is in the process of failing
func isSocketWritable(socket *socket.Socket) bool {

	engine := socket.Io().Engine()
	if engine == nil || engine.Transport() == nil {
		return false
	}

	// HasPingExpired() checks if the current ping timeout has expired
	// Writable() returns whether the transport can currently send data
	return !engine.HasPingExpired() && engine.Transport().Writable()

}
