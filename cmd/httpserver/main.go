package main

import (
	"basic_protocol/internal/headers"
	"basic_protocol/internal/request"
	"basic_protocol/internal/response"
	"basic_protocol/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func testFunc(w *response.Writer, req *request.Request) {
	var body []byte
	var statusCode response.StatusCode
	switch req.RequestTarget {
	case "/yourproblem":
		body = []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
		statusCode = response.BadRequest
	case "/myproblem":
		body = []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
		statusCode = response.InternalServerError
	default:
		body = []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
		statusCode = response.OK
	}
	h := response.GetDefaultHeaders(len(body))
	h.Override(headers.CONTENT_TYPE, "text/html")
	// log.Printf("statusLine: %s\nheaders: %+v\nbody:\n%s\n", statusCode, h, string(body))
	w.WriteStatusLine(statusCode)
	w.WriteHeaders(h)
	w.WriteBody(body)
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
