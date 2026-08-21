package bridge

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

/*
 create Unix socket
 accept Vector connection
 read Vector events
 deliver events to Agent
*/

// TODO: create event.go file
type Event []byte

var EVENT_MAX = 20_000

type Server struct {
	path string

	listener net.Listener

	event chan Event
	wg    sync.WaitGroup
}

func NewServer(path string) *Server {
	return &Server{
		path:  path,
		event: make(chan Event, EVENT_MAX),
	}
}

func (s *Server) Event() <-chan Event {
	return s.event
}

func (s *Server) acceptLoop(ctx context.Context) {

	for {

		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return

			default:
				log.Printf("accept unix socket connection: %v", err)
				continue
			}
		}

		// set other go routine to handle socket connection
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()

			s.handleConnection(ctx, conn)

		}()

	}
}

func (s *Server) Start(ctx context.Context) error {

	// Remove an old socket left behind by a previous process.
	if err := os.Remove(s.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove existing unix socket: %w", err)
	}

	listener, err := net.Listen("unix", s.path)
	if err != nil {
		return fmt.Errorf("listen on unix socket: %w", err)
	}

	// set exact permission to socket
	os.Chmod(s.path, 0660)

	s.listener = listener

	s.wg.Add(1)

	// Start a go routine to accept vector socket connection
	go func() {
		defer s.wg.Done()

		s.acceptLoop(ctx)

	}()

	return nil
}
