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

const (
	vectorStopGrace  = 15 * time.Second
	vectorStopMargin = 5 * time.Second
)

type vectorProcess struct {
	cmd      *exec.Cmd          // vector child process
	cancel   context.CancelFunc // request graceful shutdown
	done     chan struct{}      // signals that process monitoring and cleanup have completed
	stopping bool               // Guarded by Vector.mu, records weather agentv requested a graceful shutdown of the vector process
}

type Vector struct {
	Binary string         // path to the vector binary
	Config string         // path to the vector configuration file
	Agent  *agentv.Agentv // agent runtime state used to report vector status

	mu          sync.RWMutex
	lifecycleMu sync.Mutex
	running     bool
	current     *vectorProcess

	// Status callbacks
	onStatusChange func(status agentv.VectorStatus)
}

func (v *Vector) OnStatusChange(callback func(status agentv.VectorStatus)) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.onStatusChange = callback
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

func (v *Vector) startLocked(ctx context.Context) error {

	if ctx == nil {
		return fmt.Errorf("start vector: nil context")
	}
	v.mu.Lock()
	if v.current != nil {
		v.mu.Unlock()
		return fmt.Errorf("vector already running")
	}
	v.mu.Unlock()

	processCtx, cancel := context.WithCancel(ctx)
	// cmd := exec.CommandContext(processCtx, v.Binary, "--config", v.Config)
	cmd := exec.CommandContext(processCtx, v.Binary, "--config", v.Config)

	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}

		return cmd.Process.Signal(os.Interrupt)
	}
	cmd.WaitDelay = vectorStopGrace + vectorStopMargin

	p := &vectorProcess{
		cmd:    cmd,
		cancel: cancel,
		done:   make(chan struct{}),
	}

	v.mu.Lock()
	v.current = p
	v.mu.Unlock()

	// Start the vector process
	// if error occurs, cancel the context and reset the current process
	// set v.current to nil if it is still the same process
	if err := cmd.Start(); err != nil {
		cancel()
		v.mu.Lock()
		if v.current == p {
			v.current = nil
		}
		v.mu.Unlock()
		v.setStatus(agentv.VectorError)
		return fmt.Errorf("start vector: %w", err)
	}

	v.setStatus(agentv.VectorRunning)
	log.Printf("Vector started with PID: %d", cmd.Process.Pid)
	go v.monitorProcess(ctx, p)
	return nil

}

// Start runs Vector in the background
func (v *Vector) Start(ctx context.Context) error {
	v.lifecycleMu.Lock()
	defer v.lifecycleMu.Unlock()
	return v.startLocked(ctx)
}

func (v *Vector) setStatus(status agentv.VectorStatus) {
	v.Agent.Vector = status

	v.mu.RLock()
	callback := v.onStatusChange
	v.mu.RUnlock()

	if callback != nil {
		// send changed status to backend
		callback(status)
	}
}

func (v *Vector) monitorProcess(parentCtx context.Context, p *vectorProcess) {
	waitErr := p.cmd.Wait()
	p.cancel()

	v.mu.Lock()
	if v.current != p {
		// This should be rare with serialized lifecycle operations, but the
		// identity check prevents an old monitor from touching a replacement.
		v.mu.Unlock()
		close(p.done)
		return
	}
	expectedStop := p.stopping || parentCtx.Err() != nil
	v.mu.Unlock()

	switch {
	case waitErr != nil && !expectedStop:
		log.Printf("Vector crashed: %v", waitErr)
		v.setStatus(agentv.VectorCrashed)
	case waitErr != nil:
		log.Printf("Vector stopped after shutdown request: %v", waitErr)
		v.setStatus(agentv.VectorStopped)
	default:
		log.Println("Vector exited normally")
		v.setStatus(agentv.VectorStopped)
	}

	// Do not allow a replacement until final status publication is complete.
	v.mu.Lock()
	if v.current == p {
		v.current = nil
	}
	v.mu.Unlock()
	close(p.done)
}

func (v *Vector) stopLocked() error {
	v.mu.Lock()
	p := v.current
	if p == nil {
		v.mu.Unlock()
		return nil
	}
	p.stopping = true
	v.mu.Unlock()

	v.setStatus(agentv.VectorStopping)
	log.Println("Stopping Vector with SIGINT")
	p.cancel() // process cancelation

	timer := time.NewTimer(vectorStopGrace + vectorStopMargin)
	defer timer.Stop()

	select {
	case <-p.done:
		return nil
	case <-timer.C:
		return fmt.Errorf("timed out waiting for Vector to exit")
	}
}

func (v *Vector) Stop() error {
	v.lifecycleMu.Lock()
	defer v.lifecycleMu.Unlock()
	return v.stopLocked()
}

func (v *Vector) Restart(ctx context.Context) error {
	v.lifecycleMu.Lock()
	defer v.lifecycleMu.Unlock()

	if err := v.stopLocked(); err != nil {
		return fmt.Errorf("stop Vector for restart: %w", err)
	}
	// No sleep now because we want to restart as quickly as possible
	// and the stopLocked() function already waits for the process to exit.
	return v.startLocked(ctx)
}

func (v *Vector) IsRunning() bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.current != nil
}
