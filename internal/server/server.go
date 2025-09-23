package server

import (
	"fmt"
	"github.com/httpfromtcp/internal/request"
	"github.com/httpfromtcp/internal/response"
	"log"
	"net"
)

type Handler func(w *response.Writer, r *request.Request)

type Server struct {
	listener net.Listener
	handler  Handler
}

type HandleError struct {
	Status  response.StatusCode
	Message string
}

func Serve(port uint16, handler Handler) (*Server, error) {
	lst, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: lst,
		handler:  handler,
	}

	server.Listen()
	return server, nil
}

func (s *Server) Close() {
	s.listener.Close()
}

func (s *Server) Listen() {
	for {
		var err error
		con, err := s.listener.Accept()

		if err != nil {
			log.Fatal(err)
		}

		go s.handle(con)
	}
}

func (s *Server) handle(con net.Conn) {
	defer con.Close()
	req, err := request.RequestFromReader(con)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s, %s\n", req.RequestLine.RequestTarget, req.RequestLine.Method)
	writer := response.NewWriter(con)
	s.handler(writer, req)
}
