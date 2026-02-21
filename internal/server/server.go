package server

import (
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"MyOwnHTTP/internal/request"
	"MyOwnHTTP/internal/response"
)

type Server struct {
	Up        atomic.Bool
	ConnCount atomic.Int32
	Listener  net.Listener
	Handler   Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	server := Server{}
	server.Handler = handler
	listener, err := net.Listen("tcp", "localhost:"+strconv.Itoa(port))
	if err != nil {
		log.Fatalf("Failed to create a listener on the given address: %v\n", err)
	}
	server.Listener = listener
	go server.listen()
	return &server, nil
}

func (s *Server) Close() error {
	err := s.Listener.Close()
	if err != nil && s.Up.Load() {
		return err
	}
	s.Up.Store(false)
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			log.Printf("Not able to accept the incoming request: %v\n", err)
		}
		s.ConnCount.Store(s.ConnCount.Load() + 1)
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	currRequest, err := request.RequestFromReader(conn)
	if err != nil {
		log.Fatalf("Failed to read reqeust: %v\n", err)
	}

	writer := response.NewWriter(conn)
	s.Handler(writer, currRequest)
	log.Println("Response successfully sent")
	// err = conn.Write(writer.Buffer)

	s.ConnCount.Store(s.ConnCount.Load() - 1)
}
