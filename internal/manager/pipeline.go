package manager

import (
	"context"
	"fmt"

	"github.com/zakaria-jaddad/agentv/internal/bridge"
)

// NOTE: THIS IS JUST A DEMO IMPLEMENTATION OF THE DATA PIPELINE
type Sender interface {
	SendEvent(event bridge.Event) error
}

type PipeLine struct {
	input  <-chan bridge.Event
	sender Sender
}

func newPipeLine(input <-chan bridge.Event, sender Sender) *PipeLine {
	return &PipeLine{
		input:  input,
		sender: sender,
	}
}

// function that pipeline the event data
func (p *PipeLine) Run(ctx context.Context) error {

	for {
		select {
		case <-ctx.Done():
		case data, ok := <-p.input:
			if !ok {
				return nil
			}

			// TODO: Add State machine if data is failed to reach the server
			if err := p.sender.SendEvent(data); err != nil {
				return fmt.Errorf("unable to send data to backend: %v", err)
			}
		}
	}

}
