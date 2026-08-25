package bridge

import (
	"context"
	"errors"
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

var EVENT_MAX = 20_000

type Server struct {
	path string

	listener net.Listener

	event    chan EventData
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
	stopOnce sync.Once
}

func NewServer(path string) *Server {
	return &Server{
		path:  path + "vector.sock",
		event: make(chan EventData, EVENT_MAX),
	}
}

func (s *Server) Event() <-chan EventData {
	return s.event
}

func (s *Server) acceptLoop(ctx context.Context) {

	for {

		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) || ctx.Err() != nil {
				return
			}
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
		go func(c net.Conn) {
			defer s.wg.Done()

			s.handleConnection(ctx, c)

		}(conn)

	}
}

func (s *Server) Start(ctx context.Context) error {
	s.ctx, s.cancel = context.WithCancel(ctx)

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
		s.acceptLoop(s.ctx)
	}()

	return nil
}

func (s *Server) Stop() error {
	var err error
	s.stopOnce.Do(func() {
		log.Println("Closing unix socket")

		if s.cancel != nil {
			s.cancel()
		}

		if s.listener != nil {
			if closeErr := s.listener.Close(); closeErr != nil && !errors.Is(closeErr, net.ErrClosed) {
				log.Printf("error closing unix socket listener: %v", closeErr)
			}
		}

		s.wg.Wait()

		if remErr := os.Remove(s.path); remErr != nil && !os.IsNotExist(remErr) {
			log.Printf("remove existing unix socket: %v", remErr)
		}

		close(s.event)

		log.Println("Unix socket closed")
	})

	return err
}
