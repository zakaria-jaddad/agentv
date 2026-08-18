package manager

import (
	"log"

	"github.com/zakaria-jaddad/agentv/internal/agentv"
	"github.com/zakaria-jaddad/agentv/internal/config"
)

type Manager struct {
	Agent  *agentv.Agentv
	Config *config.Config
}

func New(confpath string) *Manager {

	conf, err := config.Load(confpath)
	if err != nil {
		log.Fatal(err)
	}

	if err := conf.Validate(); err != nil {
		log.Fatal(err)
	}

	agentv := agentv.New(conf.Agent.Name)
	agentv.DiscoverSystemInfo()

	return &Manager{Agent: agentv, Config: conf}
}
