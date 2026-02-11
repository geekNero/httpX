package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"basic_protocol/internal/headers"
	"basic_protocol/internal/request"
	"basic_protocol/internal/response"
	"basic_protocol/internal/server"
)

const port = 42069

func get400() ([]byte, response.StatusCode) {
	body := []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
	statusCode := response.StatusBadRequest
	return body, statusCode
}

func get500() ([]byte, response.StatusCode) {
	body := []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
	statusCode := response.StatusInternalServerError
	return body, statusCode
}

func failedTrailers(err string) headers.Headers {
	h := headers.Headers{}
	// "x-content-sha256, x-content-length, x-error"
	h.Set("x-content-sha256", "")
	h.Set("x-content-length", "")
	h.Set("x-error", err)

	return h
}

func testFunc(w *response.Writer, req *request.Request) {
	var body []byte
	var statusCode response.StatusCode
	var contentType string

	baseURL := "https://httpbin.org/"

	target, _ := strings.CutPrefix(req.RequestTarget, "/")
	target, remaining, _ := strings.Cut(target, "/")

	switch target {
	case "yourproblem":
		body, statusCode = get400()
		contentType = "text/html"
	case "myproblem":
		body, statusCode = get500()
		contentType = "text/html"
	case "video":
		var err error
		body, err = os.ReadFile("assets/vim.mp4")
		statusCode = response.StatusOK
		if err != nil {
			body, statusCode = get500()
		}
		contentType = "video/mp4"

	case "httpbin":
		// proxying for httpbin
		resp, err := http.Get(baseURL + remaining)
		defer resp.Body.Close()
		// if is not nil, it means we did not receive a standard http error from the server. We are free
		// to decide what error code to send in this case.
		if err != nil {
			log.Println("error while forwarding request, error: ", err.Error())
			body, statusCode = get500()
			break
		} else if resp.StatusCode != 200 { // if status code wasn't 200, we can simply forward the server's error as it is.
			statusCode = response.StatusCode(resp.StatusCode)
			body = []byte(resp.Status)
			break
		}

		// from here, we do not want to exit the switch case in any scenarios, as we'll be setting headers for chunk encoding
		// and the headers might be already sent to the user at the time of error. For now, we are not adding post body
		// headers, so we'll simply end the response as if it was intended.
		w.WriteStatusLine(response.StatusOK)
		h := response.GetDefaultHeaders(0) // for 0, no content-length header is added.
		h.Set("transfer-encoding", "chunked")
		if resp.Header.Get(headers.CONTENT_TYPE) != "" {
			h.Override(headers.CONTENT_TYPE, resp.Header.Get(headers.CONTENT_TYPE))
		}
		h.Set(headers.TRAILERS, "x-content-sha256, x-content-length, x-error")
		w.WriteHeaders(h)
		buffer := make([]byte, 1024)
		completeBody := []byte{}
		for { // We already have the entire body, but for the sake of learning, we'll be chunk encoding our resonse.
			n, err := resp.Body.Read(buffer)
			if err != nil {
				log.Println("failed to read response body, error: ", err.Error())
				w.WriteChunkedBodyDone(failedTrailers(err.Error()))
				return
			}
			completeBody = append(completeBody, buffer...)
			_, err = w.WriteChunkedBody(buffer[:n])
			if err != nil {
				log.Println("failed to write chunked body, error: ", err.Error())
				w.WriteChunkedBodyDone(failedTrailers(err.Error()))
				return
			}
			if n < 1024 {
				break
			}
		}

		trailers := headers.Headers{}
		hash := sha256.Sum256(completeBody)
		trailers.Set("x-content-sha256", fmt.Sprintf("%x", hash))
		trailers.Set("x-content-length", strconv.Itoa(len(completeBody)))
		trailers.Set("x-error", "")

		w.WriteChunkedBodyDone(trailers)
		return

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
		statusCode = response.StatusOK
	}
	h := response.GetDefaultHeaders(len(body))
	h.Override(headers.CONTENT_TYPE, contentType)
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
