package backend

import "github.com/zakaria-jaddad/agentv/internal/bridge"

func (c *Client) SendEvent(data bridge.Event) error {
	return c.Emit("agent:data", data)
}
