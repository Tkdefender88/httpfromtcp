package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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
		handler500(w, req)
		return
	}
	defer resp.Body.Close()

	h := headers.New()
	h.Set("Transfer-Encoding", "chunked")
	h.Set("Content-Type", resp.Header.Get("Content-Type"))
	h.Set("Trailer", "X-Content-SHA256")
	h.Set("Trailer", "X-Content-Length")

	w.WriteStatus(response.StatusOK)
	w.WriteHeaders(h)

	responseBody := make([]byte, 0, 1024)
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			w.WriteChunkedBody(buf[:n])
			responseBody = append(responseBody, buf[:n]...)
		}
		if err != nil {
			if errors.Is(io.EOF, err) {
				break
			}
			fmt.Printf("error reading response: %v", err)
			break
		}
	}

	hash := sha256.Sum256(responseBody)
	w.WriteChunkedBodyDone()
	trailers := headers.New()
	trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", hash[:]))
	trailers.Set("X-Content-Length", strconv.Itoa(len(responseBody)))
	w.WriteTrailers(trailers)
}

func handlerVideo(w *response.Writer, req *request.Request) {
	h := headers.New()
	h.Set("Content-Type", "video/mp4")

	data, err := os.ReadFile("./assets/vim.mp4")
	if err != nil {
		fmt.Printf("error occurred: %v\n", err)
		handler500(w, req)
		return
	}
	h.Set("Content-Length", strconv.Itoa(len(data)))

	w.WriteStatus(response.StatusOK)
	w.WriteHeaders(h)
	w.WriteBody(data)
}

func handlerFunc(w *response.Writer, req *request.Request) {
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		handleProxy(w, req)
		return
	}

	switch req.RequestLine.RequestTarget {
	case "/video":
		handlerVideo(w, req)
	case "/yourproblem":
		handler400(w, req)
	case "/myproblem":
		handler500(w, req)
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

func handler400(w *response.Writer, _ *request.Request) {
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
}

func handler500(w *response.Writer, _ *request.Request) {
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
