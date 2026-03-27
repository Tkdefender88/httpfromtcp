package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

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

func handlerFunc(w *response.Writer, req *request.Request) {
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
