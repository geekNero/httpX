// Package server provides a server instance that listens to HTTP 1.1 requests on a TCP connections at the provided port and returns a valid HTTP response
package server

import (
	"basic_protocol/internal/common"
	"errors"
	"log"
	"net"
	"strconv"
	"sync/atomic"
)

var (
	temp = []byte(`HTTP/1.1 200 OK
Content-Type: text/plain
Content-Length: 13

Hello World!\n`)
)

type Server struct {
	port     int
	running  atomic.Bool
	listener net.Listener
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen(common.TCP, common.LocalHost+":"+strconv.Itoa(port))

	server := Server{}
	if err != nil {
		return nil, err
	}
	server.listener = listener
	server.running.Store(true)
	go server.listen()
	return &server, nil
}

func (s *Server) handle(conn net.Conn) {
	_, err := conn.Write(temp)
	if err != nil {
		log.Println("failed to write to connection, error: ", err.Error())
	}
	conn.Close()
	log.Println("Written to connection")
}

func (s *Server) listen() {
	for s.IsRunning() {
		log.Println("Waiting on connection")
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				log.Println("Server closed, stop accepting new connections")
				return
			}
			log.Println("error", err.Error())
			continue
		}
		log.Println("Connection established")
		go s.handle(conn)
	}
}

func (s *Server) Close() error {
	s.running.Store(false)
	err := s.listener.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) IsRunning() bool {
	return s.running.Load()
}
