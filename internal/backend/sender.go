package backend

import (
	"fmt"

	"github.com/zakaria-jaddad/agentv/internal/agentv"
	"github.com/zakaria-jaddad/agentv/internal/bridge"
)

func (c *Client) SendEvent(event string, data bridge.EventData) error {
	return c.Emit(event, string(data))
}

func (c *Client) SendStatusUpdate(status agentv.VectorStatus) error {

	if err := c.Emit(string(bridge.VectorStatus), status); err != nil {
		return fmt.Errorf("emit failed: %v", err)
	}
	return nil
}
