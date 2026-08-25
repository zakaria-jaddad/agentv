package manager

import (
	"log"
	"os"

	"github.com/zakaria-jaddad/agentv/internal/bridge"
)

// RegisterEvents(): Register and Handle incoming socket events
// TODO: After a successful event, agent should Emit the same event the to backend indicating a job is done
func (m *Manager) RegisterEvents() error {

	// Event registration for "vector:restart"
	m.BackendClient.RegisterEvent(bridge.VectorRestart, func() error {

		log.Println("Restarting Vector...")
		err := m.RestartVector()
		if err != nil {
			return err
		}
		log.Println("Vector Restarted Successfully")
		return nil
	})

	// Event registration for "vector:stop"
	m.BackendClient.RegisterEvent(bridge.VectorStop, func() error {

		log.Println("Stopping Vector...")
		err := m.StopVector()
		if err != nil {
			return err
		}
		log.Println("Vector Stopped Successfully")
		return nil
	})

	// Event registration for "vector:start"
	m.BackendClient.RegisterEvent(bridge.VectorStart, func() error {

		log.Println("Starting Vector...")
		err := m.StartVector()
		if err != nil {
			return err
		}
		log.Println("Vector Started Successfully")
		return nil
	})

	// Event registration for "vector:stop"
	m.BackendClient.RegisterEvent(bridge.AgentStop, func() error {

		err := m.Stop()
		if err != nil {
			return err
		}
		log.Println("Exiting...")
		os.Exit(0)
		return nil
	})

	return nil
}
