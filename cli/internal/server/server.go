package server

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/lossystyles/cli/internal/protocol"
)

type Server struct {
	listener net.Listener
	sockPath string
	Messages chan protocol.Message
	done     chan struct{}
}

func SocketPath(runID string) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("lossystyles-%s.sock", runID))
}

func New(sockPath string) (*Server, error) {
	// Clean up stale socket
	os.Remove(sockPath)

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", sockPath, err)
	}

	return &Server{
		listener: listener,
		sockPath: sockPath,
		Messages: make(chan protocol.Message, 256),
		done:     make(chan struct{}),
	}, nil
}

func (s *Server) Accept() {
	defer close(s.Messages)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.done:
				return
			default:
				continue
			}
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	// Allow large messages (1MB)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		msg, err := protocol.Parse(scanner.Bytes())
		if err != nil {
			continue
		}
		s.Messages <- msg
	}
}

func (s *Server) Close() error {
	close(s.done)
	s.listener.Close()
	return os.Remove(s.sockPath)
}
