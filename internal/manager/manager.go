package manager

import (
	"github.com/zakaria-jaddad/agentv/internal/agentv"
	"github.com/zakaria-jaddad/agentv/internal/backend"
	"github.com/zakaria-jaddad/agentv/internal/bridge"
	"github.com/zakaria-jaddad/agentv/internal/config"
)

type Manager struct {
	Agent         *agentv.Agentv  // agent runtime information
	Config        *config.Config  // agent configuration
	BackendClient *backend.Client // backend client used to authenticate and send data with socket io
	BridgeServer  *bridge.Server  // unix socket for communication between vector and agentv
	Pipeline      *PipeLine       // vector data pipeline to the backend
}

func New(
	agentv *agentv.Agentv,
	conf *config.Config,
	backendClient *backend.Client,
	bridgeServer *bridge.Server,
) *Manager {

	pipeline := newPipeLine(bridgeServer.Event(), backendClient)
	return &Manager{
		Agent:         agentv,
		Config:        conf,
		BackendClient: backendClient,
		BridgeServer:  bridgeServer,
		Pipeline:      pipeline,
	}
}
