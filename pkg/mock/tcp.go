package mock

import (
	"io"
	"net"

	"github.com/alauda/kube-supv/pkg/log"
)

type TCPHandler func(net.Conn)

type TCPServer struct {
	addr     string
	listener net.Listener
	stop     chan bool
	handler  TCPHandler
}

func NewTCPServer(addr string, handler TCPHandler) (io.Closer, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	s := &TCPServer{
		addr:     addr,
		listener: listener,
		stop:     make(chan bool),
	}
	go s.run()
	return s, nil
}

func (s *TCPServer) run() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.stop:
				return
			default:
				log.Errorf(`listen "%s", error: %v`, s.addr, err)
			}
		}
		go s.handler(conn)
	}
}

func (s *TCPServer) Close() error {
	close(s.stop)
	return s.listener.Close()
}
