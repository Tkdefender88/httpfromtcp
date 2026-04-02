package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Tkdefender88/httpfromtcp/internal/headers"
	"github.com/Tkdefender88/httpfromtcp/internal/request"
	"github.com/Tkdefender88/httpfromtcp/internal/response"
	"github.com/Tkdefender88/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handlerFunc)
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

func handleProxy(w *response.Writer, req *request.Request) {
	target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")

	resp, err := http.Get(fmt.Sprintf("https://httpbin.org/%s", target))
	if err != nil {
		fmt.Printf("error proxying request: %v\n", err)
		return
	}

	h := headers.New()
	h.Set("Transfer-Encoding", "chunked")
	h.Set("Content-Type", resp.Header.Get("Content-Type"))

	w.WriteStatus(response.StatusOK)
	w.WriteHeaders(h)

	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			w.WriteChunkedBody(buf[:n])
		}
		if err != nil {
			if errors.Is(io.EOF, err) {
				break
			}
			fmt.Printf("error reading response: %v", err)
			break
		}
	}

	w.WriteChunkedBodyDone()
}

func handlerFunc(w *response.Writer, req *request.Request) {
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		handleProxy(w, req)
		return
	}

	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		body := "<html>\n" +
			"<head>\n" +
			"<title>400 Bad Request</title>\n" +
			"</head>\n" +
			"<body>\n" +
			"<h1>Bad Request</h1>\n" +
			"<p>Your request honestly kinda sucked.</p>\n" +
			"</body>\n" +
			"</html>\n"
		WriteHTML(w, response.StatusBadRequest, []byte(body))
	case "/myproblem":
		body := "<html>\n" +
			"<head>\n" +
			"<title>500 Internal Server Error</title>\n" +
			"</head>\n" +
			"<body>\n" +
			"<h1>Internal Server Error</h1>\n" +
			"<p>Okay, you know what? This one is on me.</p>\n" +
			"</body>\n" +
			"</html>\n"
		WriteHTML(w, response.StatusInternalServerError, []byte(body))
	default:
		body := "<html>\n" +
			"<head>\n" +
			"<title>200 OK</title>\n" +
			"</head>\n" +
			"<body>\n" +
			"<h1>Success!</h1>\n" +
			"<p>Your request was an absolute banger.</p>\n" +
			"</body>\n" +
			"</html>\n"
		WriteHTML(w, response.StatusOK, []byte(body))
	}
}

func WriteHTML(w *response.Writer, statusCode response.StatusCode, body []byte) {
	h := response.SetDefaultHeaders(len(body))
	h.Set("Content-Type", "text/html")
	w.WriteStatus(statusCode)
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func WriteText(w *response.Writer, statusCode response.StatusCode, body []byte) {
	h := response.SetDefaultHeaders(len(body))
	h.Set("Content-Type", "text/plain")
	w.WriteStatus(statusCode)
	w.WriteHeaders(h)
	w.WriteBody(body)
}
