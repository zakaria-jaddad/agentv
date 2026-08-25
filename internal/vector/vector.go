package vector

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/zakaria-jaddad/agentv/internal/agentv"
)

type Vector struct {
	Binary string         // path to the vector binary
	Config string         // path to the vector configuration file
	Agent  *agentv.Agentv // agent runtime state used to report vector status

	// Vector sub process management
	cmd      *exec.Cmd
	mu       sync.RWMutex
	running  bool
	cancel   context.CancelFunc
	stopChan chan struct{}
	doneChan chan error

	// Status callbacks
	onStatusChange func(status agentv.VectorStatus)
}

func (v *Vector) OnStatusChange(callback func(status agentv.VectorStatus)) {
	v.onStatusChange = callback
}

type VectorStatus int

func New(binary string, config string, agent *agentv.Agentv) *Vector {
	return &Vector{
		Binary:   binary,
		Config:   config,
		Agent:    agent,
		doneChan: make(chan error, 1),
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

// Start runs Vector in the background
func (v *Vector) Start(ctx context.Context) error {

	v.mu.Lock()
	if v.running {
		v.mu.Unlock()
		return fmt.Errorf("vector already running")
	}

	ctx, cancel := context.WithCancel(ctx)
	v.cancel = cancel
	v.stopChan = make(chan struct{})

	cmd := exec.CommandContext(ctx, v.Binary, "--config", v.Config)
	v.cmd = cmd
	v.running = true
	v.mu.Unlock()

	if err := cmd.Start(); err != nil {
		v.mu.Lock()
		v.running = false
		v.mu.Unlock()
		v.setStatus(agentv.VectorError)
		return fmt.Errorf("start vector: %w", err)
	}
	v.setStatus(agentv.VectorRunning)
	log.Printf("Vector started with PID: %d", cmd.Process.Pid)

	go v.monitorProcess(cmd)

	return nil
}

func (v *Vector) setStatus(status agentv.VectorStatus) {
	v.Agent.Vector = status
	if v.onStatusChange != nil {
		// send changed status to backend
		v.onStatusChange(status)
	}
}

func (v *Vector) monitorProcess(cmd *exec.Cmd) {
	err := cmd.Wait()
	v.mu.Lock()
	defer v.mu.Unlock()

	v.running = false

	if err != nil {
		select {
		case <-v.stopChan:
			v.setStatus(agentv.VectorStopped)
			log.Println("Vector stopped gracefully")
		default:
			v.setStatus(agentv.VectorCrashed)
			log.Printf("vector crashed: %v", err)
		}
	} else {
		v.setStatus(agentv.VectorStopped)
		log.Println("Vector exited normally")
	}

	select {
	case v.doneChan <- err:
	default:
	}
}

func (v *Vector) Stop() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if !v.running {
		return nil
	}

	v.setStatus(agentv.VectorStopping)
	log.Printf("Stopping Vector...")

	if v.stopChan != nil {
		select {
		case <-v.stopChan:
		default:
			close(v.stopChan)
		}
	}

	if v.cancel != nil {
		v.cancel()
	}

	// stop process by sending sigint or hard kill
	if v.cmd != nil && v.cmd.Process != nil {
		if err := v.cmd.Process.Signal(os.Interrupt); err != nil {
			// kill process using kill syscall
			log.Printf("SIGINT failed: %v, forcing kill", err)
			if err := v.cmd.Process.Kill(); err != nil {
				return fmt.Errorf("kill vector: %w", err)
			}
		}
	}

	v.running = false
	return nil
}

func (v *Vector) Restart(ctx context.Context) error {

	// stop vector
	// ignore if vector already stopping
	if err := v.Stop(); err != nil && err.Error() != "vector is not running" {
		return fmt.Errorf("stop for vector restart: %w", err)
	}

	// little time out for process clean up
	time.Sleep(2 * time.Second)

	// resetting done channel
	v.mu.Lock()
	v.doneChan = make(chan error, 1)
	v.mu.Unlock()

	return v.Start(ctx)
}

func (v *Vector) IsRunning() bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.running
}
