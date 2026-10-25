package manager

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/zakaria-jaddad/agentv/internal/backend"
	"github.com/zakaria-jaddad/agentv/internal/bridge"
)

const backendRetryDelay = time.Second

type Sender interface {
	SendEvent(event string, data bridge.EventData) error
}

type PipeLine struct {
	input  <-chan bridge.EventData
	sender Sender
}

func newPipeLine(input <-chan bridge.EventData, sender Sender) *PipeLine {
	return &PipeLine{
		input:  input,
		sender: sender,
	}
}

// function that pipeline the event data
func (p *PipeLine) Run(ctx context.Context) {

	for {
		select {
		case <-ctx.Done():
			fmt.Println("pipeline go routine is done stopping")
			return
		case data, ok := <-p.input:
			if !ok {
				return
			}

			// send the data to the backend with retry, return error when unexpected error, return when backend is unavailable
			if err := p.sendWithRetry(ctx, data); err != nil {
				log.Printf("unable to send data to backend: %v", err)
			}
		}
	}
}

func (p *PipeLine) sendWithRetry(ctx context.Context, data bridge.EventData) error {
	for {
		err := p.sender.SendEvent(string(bridge.AgentData), data)
		if err == nil {
			return nil
		}

		// if the error is not backend unavailable, return the error
		if !errors.Is(err, backend.ErrBackendUnavailable) {
			return err
		}

		timer := time.NewTimer(backendRetryDelay)

		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()

		// hang here untile the backendRetryDelay is over
		case <-timer.C:
		}
	}
}
