// Package server provides a server instance that listens to HTTP 1.1 requests on a TCP connections at the provided port and returns a valid HTTP response
package server

import (
	"basic_protocol/internal/common"
	"basic_protocol/internal/request"
	"basic_protocol/internal/response"
	"errors"
	"log"
	"net"
	"strconv"
	"sync/atomic"
)

type Server struct {
	port     int
	running  atomic.Bool
	listener net.Listener
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	server := Server{}
	listener, err := net.Listen(common.TCP, common.LocalHost+":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	server.handler = handler
	server.listener = listener
	server.running.Store(true)
	go server.listen()
	return &server, nil
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	resp := response.NewResponseWriter(conn)
	request, err := request.RequestFromReader(conn)
	if err != nil {
		errorString := "failed to parse request, error: " + err.Error()
		err1 := resp.WriteStatusLine(response.BadRequest)
		err2 := resp.WriteHeaders(response.GetDefaultHeaders(len(errorString)))
		_, err = resp.WriteBody([]byte(errorString))

		if err != nil || err1 != nil || err2 != nil {
			log.Printf("failure while writing bad request to response, error: statusLine: %s, headers: %s, body: %s\n", err1.Error(), err2.Error(), err.Error())
		}
		return
	}
	s.handler(resp, request)
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
