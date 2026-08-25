package manager

import (
	"context"
	"fmt"

	"log"

	"github.com/zakaria-jaddad/agentv/internal/bridge"
)

// NOTE: THIS IS JUST A DEMO IMPLEMENTATION OF THE DATA PIPELINE
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

			// TODO: Add State machine if data is failed to reach the server
			if err := p.sender.SendEvent(string(bridge.AgentData), data); err != nil {
				log.Printf("unable to send data to backend: %v", err)
			}
		}
	}

}
