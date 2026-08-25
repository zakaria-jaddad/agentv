package vector

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/zakaria-jaddad/agentv/internal/agentv"
)

func TestVectorStartStop(t *testing.T) {
	// Find sleep command
	sleepPath, err := exec.LookPath("sleep")
	if err != nil {
		t.Skip("sleep binary not found")
	}

	agent := agentv.New("test-agent")
	vec := New(sleepPath, "10", agent)
	// Override command in Start by creating custom or testing lifecycle
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vec.mu.Lock()
	vec.stopChan = make(chan struct{})
	cmdCtx, cmdCancel := context.WithCancel(ctx)
	vec.cancel = cmdCancel
	cmd := exec.CommandContext(cmdCtx, sleepPath, "5")
	vec.cmd = cmd
	vec.running = true
	vec.mu.Unlock()

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start test process: %v", err)
	}
	go vec.monitorProcess(cmd)

	if !vec.IsRunning() {
		t.Fatalf("expected vector to be running")
	}

	// Test Stop
	done := make(chan error, 1)
	go func() {
		done <- vec.Stop()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Stop returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("vec.Stop hung!")
	}

	// Test stopping already stopped vector
	if err := vec.Stop(); err != nil {
		t.Errorf("second Stop() returned error: %v", err)
	}
}

func TestVectorStopWhenNotRunning(t *testing.T) {
	agent := agentv.New("test-agent")
	vec := New("/bin/dummy", "config.yaml", agent)

	if err := vec.Stop(); err != nil {
		t.Errorf("Stop on unstarted vector should not error, got: %v", err)
	}
}
