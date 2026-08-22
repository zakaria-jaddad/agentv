package manager

import (
	"context"

	"github.com/zakaria-jaddad/agentv/internal/agentv"
	"github.com/zakaria-jaddad/agentv/internal/backend"
	"github.com/zakaria-jaddad/agentv/internal/bridge"
	"github.com/zakaria-jaddad/agentv/internal/config"
	"github.com/zakaria-jaddad/agentv/internal/vector"
)

type Manager struct {
	Agent         *agentv.Agentv  // agent runtime information
	Config        *config.Config  // agent configuration
	BackendClient *backend.Client // backend client used to authenticate and send data with socket io
	BridgeServer  *bridge.Server  // unix socket for communication between vector and agentv
	Vector        *vector.Vector  // vector process lifecycle: config validation, run, status
	Pipeline      *PipeLine       // vector data pipeline to the backend
}

func New(
	agentv *agentv.Agentv,
	conf *config.Config,
	backendClient *backend.Client,
	bridgeServer *bridge.Server,
) *Manager {

	pipeline := newPipeLine(bridgeServer.Event(), backendClient)
	vec := vector.New(conf.Vector.Binary, conf.Vector.Config, agentv)
	return &Manager{
		Agent:         agentv,
		Config:        conf,
		BackendClient: backendClient,
		BridgeServer:  bridgeServer,
		Vector:        vec,
		Pipeline:      pipeline,
	}
}

func (m *Manager) Run(ctx context.Context) error {
	return m.Pipeline.Run(ctx)
}
