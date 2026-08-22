package vector

import (
	"context"
	"fmt"
	"log"
	"os/exec"

	"github.com/zakaria-jaddad/agentv/internal/agentv"
)

type Vector struct {
	Binary string         // path to the vector binary
	Config string         // path to the vector configuration file
	Agent  *agentv.Agentv // agent runtime state used to report vector status
}

func New(binary string, config string, agent *agentv.Agentv) *Vector {
	return &Vector{
		Binary: binary,
		Config: config,
		Agent:  agent,
	}
}

// ValidateConfig checks that the vector binary exists
// and validates its configuration file.
func (v *Vector) ValidateConfig() error {
	log.Println("Starting Vector Config Validation")

	if _, err := exec.LookPath(v.Binary); err != nil {
		return fmt.Errorf("vector binary not found: %w", err)
	}

	output, err := exec.Command(v.Binary, "validate", v.Config).CombinedOutput()
	if err != nil {
		return fmt.Errorf("vector config validation failed: %w\n\tOutput: %s", err, output)
	}

	log.Println("Vector Config Validation Done")
	if len(output) > 0 {
		log.Printf("Vector validation output: %s", output)
	}

	return nil
}

// Start runs Vector in the foreground and blocks until it exits.
func (v *Vector) Start(ctx context.Context) error {
	v.Agent.Vector = agentv.VectorStarting

	cmd := exec.CommandContext(ctx, v.Binary, "--config", v.Config)

	if err := cmd.Start(); err != nil {
		v.Agent.Vector = agentv.VectorError
		return fmt.Errorf("start vector: %w", err)
	}
	v.Agent.Vector = agentv.VectorRunning

	if err := cmd.Wait(); err != nil {
		v.Agent.Vector = agentv.VectorCrashed
		return fmt.Errorf("vector exited: %w", err)
	}

	v.Agent.Vector = agentv.VectorStopped
	return nil
}
