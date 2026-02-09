package main

import (
	"basic_protocol/internal/request"
	"basic_protocol/internal/response"
	"basic_protocol/internal/server"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func testFunc(w io.Writer, req *request.Request) *server.HandlerError {
	var herr *server.HandlerError
	switch req.RequestTarget {
	case "/yourproblem":
		herr = &server.HandlerError{
			StatusCode: response.BadRequest,
			Message:    "Your problem is not my problem\n",
		}
	case "/myproblem":
		herr = &server.HandlerError{
			StatusCode: response.InternalServerError,
			Message:    "Woopsie, my bad\n",
		}
	default:
		w.Write([]byte("All good, frfr\n"))
	}
	return herr
}

func main() {
	server, err := server.Serve(port, testFunc)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Server gracefully stopped")
}
