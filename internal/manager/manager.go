package manager

import (
	"context"
	"log"
	"sync"
	"time"

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

	// Internal Manager Attributes
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	stopOnce sync.Once
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

func (m *Manager) Start(ctx context.Context) error {
	m.ctx, m.cancel = context.WithCancel(ctx)

	// Vector lifecycle: validate config then run Vector in the background
	if err := m.Vector.ValidateConfig(); err != nil {
		log.Fatalf("%v", err)
	}

	// Set up status change callbacks
	m.Vector.OnStatusChange(func(status agentv.VectorStatus) {
		log.Printf("Vector status changed to: %v", status)
		// Send status update to backend
		m.BackendClient.SendStatusUpdate(status)
	})

	// Start Vector in background
	if err := m.Vector.Start(m.ctx); err != nil {
		return err
	}

	// The mechanism relies on three primary methods:
	// Add(delta int): Increments the internal counter by a specified amount, typically called before launching a new goroutine.
	// Done(): Decrements the counter by one, usually deferred at the start of a worker goroutine to signal completion.
	// Wait(): Blocks the calling goroutine until the internal counter reaches zero, ensuring all tracked tasks have completed.
	// m.wg.Add(1)
	// go m.runBackgroundTask(m.ctx)

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.Pipeline.Run(m.ctx)
	}()

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-m.ctx.Done():
				return
			case <-ticker.C:
				// heart beat
				m.BackendClient.Emit(string(bridge.AgentHeartBeat), m.Agent.Vector)
			}
		}
	}()

	return nil
}

// a go routine for background health check
// NOTE: Should i keep this
// func (m *Manager) runBackgroundTask(ctx context.Context) {
// 	defer m.wg.Done()
// 	ticker := time.NewTicker(30 * time.Second)
// 	defer ticker.Stop()
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		case <-ticker.C:
// 			// health check
// 			if m.Vector.IsRunning() == false {
// 				// log.Println("Vector not running, attempting restart...")
// 				// if err := m.Vector.Restart(ctx); err != nil {
// 				// 	log.Printf("Failed to restart vector: %v", err)
// 			}
// 		}
// 	}
// }

// Restart restarts Vector
func (m *Manager) RestartVector() error {
	return m.Vector.Restart(m.ctx)
}

// StopVector stops Vector
func (m *Manager) StopVector() error {
	return m.Vector.Stop()
}

// StartVector starts Vector
func (m *Manager) StartVector() error {
	return m.Vector.Start(m.ctx)
}

// Stop stops all manager components
func (m *Manager) Stop() error {
	var err error
	m.stopOnce.Do(func() {
		log.Println("Stopping manager...")

		// Cancel manager context first to stop background tasks
		if m.cancel != nil {
			m.cancel()
		}

		// Stop Vector gracefully
		if m.Vector != nil {
			if stopErr := m.Vector.Stop(); stopErr != nil {
				log.Printf("Warning: error stopping vector: %v", stopErr)
			}
		}

		// Stopping unix socket
		if m.BridgeServer != nil {
			if stopErr := m.BridgeServer.Stop(); stopErr != nil {
				log.Printf("Warning: error stopping bridge server: %v", stopErr)
			}
		}

		// Wait for manager background goroutines
		m.wg.Wait()

		// Disconnect backend client socket if connected
		if m.BackendClient != nil {
			m.BackendClient.Disconnect()
		}

		log.Println("Manager stopped")
	})
	return err
}
