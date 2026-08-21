package bridge

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {

	// close the connection when function complete
	defer conn.Close()

	remoteAddress := conn.RemoteAddr().String()

	log.Printf("Vector connected to unix socket")

	reader := bufio.NewReader(conn)

	for {
		rawData, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				log.Printf("vector disconnected from %s", remoteAddress)
			} else {
				log.Printf("read error from %s: %v", remoteAddress, err)
			}
			return
		}

		// This block will never stop
		// data would be read from the socket
		// and would be transformed to the channel etc...
		select {
		case <-ctx.Done():
			fmt.Printf("closing connection from %s: ", remoteAddress)
			return
		case s.event <- rawData:
		}
	}
}

func (s *Server) Close() error {

	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return err
		}
	}

	s.wg.Wait()
	if err := os.Remove(s.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove existing unix socket: %w", err)
	}

	close(s.event)

	return nil
}
