package bridge

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log"
	"net"
)

func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {

	// close the connection when function complete
	defer conn.Close()

	remoteAddress := conn.RemoteAddr().String()

	log.Printf("Vector connected to unix socket")

	// Close conn when ctx is cancelled to unblock any blocking ReadBytes
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		select {
		case <-ctx.Done():
			conn.Close()
		case <-stop:
		}
	}()

	reader := bufio.NewReader(conn)

	for {
		rawData, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF && !errors.Is(err, net.ErrClosed) {
				log.Printf("read error from %s: %v", remoteAddress, err)
			}
			return
		}

		// This block will never stop
		// data would be read from the socket
		// and would be transformed to the channel etc...
		select {
		case <-ctx.Done():
			log.Printf("closing connection from %s", remoteAddress)
			return
		case s.event <- rawData:
		}
	}
}

func (s *Server) Close() error {
	return s.Stop()
}

