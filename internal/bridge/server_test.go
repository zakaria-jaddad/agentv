package bridge

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestServerStartStop(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bridge_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sockPath := filepath.Join(tmpDir, "test.sock")
	srv := NewServer(sockPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := srv.Start(ctx); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	// Verify socket file exists
	if _, err := os.Stat(sockPath); os.IsNotExist(err) {
		t.Fatalf("socket file does not exist")
	}

	// Stop should complete quickly and not hang
	done := make(chan error, 1)
	go func() {
		done <- srv.Stop()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("srv.Stop returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("srv.Stop hung!")
	}

	// Verify double stop is safe
	if err := srv.Stop(); err != nil {
		t.Errorf("second Stop() returned error: %v", err)
	}
}

func TestServerStopWithActiveConnection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bridge_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sockPath := filepath.Join(tmpDir, "test.sock")
	srv := NewServer(sockPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := srv.Start(ctx); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	// Connect a client that stays idle
	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatalf("failed to connect client: %v", err)
	}
	defer conn.Close()

	// Send a line
	fmt.Fprintf(conn, "test message\n")

	// Read from event channel
	select {
	case msg := <-srv.Event():
		if string(msg) != "test message\n" {
			t.Errorf("expected 'test message\\n', got %q", string(msg))
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("did not receive event")
	}

	// Client remains connected and idle; Stop() should still terminate promptly
	done := make(chan error, 1)
	go func() {
		done <- srv.Stop()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("srv.Stop returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("srv.Stop hung while client was connected!")
	}
}
