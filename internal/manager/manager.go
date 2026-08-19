package manager

import (
	"github.com/zakaria-jaddad/agentv/internal/agentv"
	"github.com/zakaria-jaddad/agentv/internal/config"
)

type Manager struct {
	Agent  *agentv.Agentv
	Config *config.Config
}

func New(confpath string) (*Manager, error) {

	conf, err := config.Load(confpath)
	if err != nil {
		return nil, err
	}

	if err := conf.Validate(); err != nil {
		return nil, err
	}

	agentv := agentv.New(conf.Agent.Name)
	err = agentv.DiscoverSystemInfo()
	if err != nil {
		return nil, err
	}

	return &Manager{Agent: agentv, Config: conf}, nil
}
