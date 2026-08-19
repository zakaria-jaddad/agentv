package agentv

import (
	"os"
	"runtime"
	"time"
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusActive   Status = "active"
	StatusDisabled Status = "disabled"
	StatusRevoked  Status = "revoked"
)

type ConnectionStatus string

const (
	ConnectionDisconnected ConnectionStatus = "disconnected"
	ConnectionConnecting   ConnectionStatus = "connecting"
	ConnectionConnected    ConnectionStatus = "connected"
)

type VectorStatus string

const (
	VectorStopped  VectorStatus = "stopped"
	VectorStarting VectorStatus = "starting"
	VectorRunning  VectorStatus = "running"
	VectorStopping VectorStatus = "stopping"
	VectorCrashed  VectorStatus = "crashed"
	VectorError    VectorStatus = "error"
)

type Agentv struct {
	ID           int64
	Name         string
	Hostname     string
	OS           string
	Architecture string

	Version       string
	VectorVersion string

	Status     Status
	Connection ConnectionStatus
	Vector     VectorStatus

	CreatedAt          time.Time
	LastSeenAt         time.Time
	LastConnectedAt    time.Time
	LastDisconnectedAt time.Time
}

func New(name string) *Agentv {
	return &Agentv{
		Name:       name,
		Status:     StatusPending,
		Connection: ConnectionConnecting,
		Vector:     VectorStopped,
		CreatedAt:  time.Now(),
	}
}

func (a *Agentv) DiscoverSystemInfo() error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	a.Hostname = hostname
	a.OS = runtime.GOOS
	a.Architecture = runtime.GOARCH

	return nil
}
