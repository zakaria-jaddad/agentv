package manager

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/zakaria-jaddad/agentv/internal/agentv"
	"github.com/zakaria-jaddad/agentv/internal/backend"
	"github.com/zakaria-jaddad/agentv/internal/bridge"
	"github.com/zakaria-jaddad/agentv/internal/config"
)

func TestManagerStopGraceful(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mgr_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sockPath := filepath.Join(tmpDir, "test.sock")
	agent := agentv.New("test-agent")
	conf := &config.Config{
		Agent:  config.AgentConfig{Name: "test-agent"},
		Vector: config.VectorConfig{Binary: "echo", Config: "", Socket: sockPath},
	}
	backendClient := backend.NewClient("http://localhost:9999", "test-token")
	bridgeServer := bridge.NewServer(sockPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := bridgeServer.Start(ctx); err != nil {
		t.Fatalf("failed to start bridge server: %v", err)
	}

	mgr := New(agent, conf, backendClient, bridgeServer)

	// Start manager background goroutines manually or via Start
	mgr.ctx, mgr.cancel = context.WithCancel(ctx)
	mgr.wg.Add(1)
	// go mgr.runBackgroundTask(mgr.ctx)

	mgr.wg.Add(1)
	go func() {
		defer mgr.wg.Done()
		mgr.Pipeline.Run(mgr.ctx)
	}()

	mgr.wg.Add(1)
	go func() {
		defer mgr.wg.Done()
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-mgr.ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()

	// Stop manager and ensure it completes within 2 seconds without hanging
	done := make(chan error, 1)
	go func() {
		done <- mgr.Stop()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("mgr.Stop returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("mgr.Stop hung!")
	}

	// Double Stop should be safe
	if err := mgr.Stop(); err != nil {
		t.Errorf("second mgr.Stop returned error: %v", err)
	}
}
