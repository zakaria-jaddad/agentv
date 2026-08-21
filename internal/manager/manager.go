package manager

import (
	"github.com/zakaria-jaddad/agentv/internal/agentv"
	"github.com/zakaria-jaddad/agentv/internal/config"
)

type Manager struct {
	Agent  *agentv.Agentv
	Config *config.Config
}

func New(agentv *agentv.Agentv, conf *config.Config) *Manager {
	return &Manager{Agent: agentv, Config: conf}
}
