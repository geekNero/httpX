// Package server provides a server instance that listens to HTTP 1.1 requests on a TCP connections at the provided port and returns a valid HTTP response
package server

import (
	"basic_protocol/internal/common"
	"basic_protocol/internal/request"
	"basic_protocol/internal/response"
	"bytes"
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
	request, err := request.RequestFromReader(conn)
	if err != nil {
		response.WriteStatusLine(conn, response.BadRequest)
		conn.Write([]byte(err.Error()))
		return
	}
	respWriter := bytes.Buffer{}
	herr := s.handler(&respWriter, request)
	var statusCode response.StatusCode
	var responseBody []byte
	if herr != nil {
		statusCode = herr.StatusCode
		responseBody = []byte(herr.Message)
	} else {
		statusCode = response.OK
		responseBody = respWriter.Bytes()
	}
	err = response.WriteStatusLine(conn, statusCode)
	if err != nil {
		log.Println("failed to write status, error", err.Error())
		return
	}

	err = response.WriteHeaders(conn, response.GetDefaultHeaders(len(responseBody)))
	if err != nil {
		log.Println("failed to write headers, error", err.Error())
		return
	}

	conn.Write(responseBody)
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
