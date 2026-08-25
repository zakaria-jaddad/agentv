package backend

import (
	"fmt"

	"github.com/zakaria-jaddad/agentv/internal/agentv"
	"github.com/zakaria-jaddad/agentv/internal/bridge"
)

func (c *Client) SendEvent(event string, data bridge.EventData) error {
	return c.Emit(event, string(data))
}

type StatusUpdate struct {
	AgentID int                 `json:"agentID"`
	Status  agentv.VectorStatus `json:"status"`
}

func (c *Client) SendStatusUpdate(agentID int, status agentv.VectorStatus) error {
	update := StatusUpdate{
		AgentID: agentID,
		Status:  status,
	}

	if err := c.Emit(string(bridge.AgentStatus), update); err != nil {
		return fmt.Errorf("emit failed: %v", err)
	}
	return nil
}
