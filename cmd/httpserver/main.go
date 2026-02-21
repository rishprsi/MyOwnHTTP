package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"MyOwnHTTP/internal/headers"
	"MyOwnHTTP/internal/request"
	"MyOwnHTTP/internal/response"
	"MyOwnHTTP/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
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

func handler(w *response.Writer, req *request.Request) {
	if req.RequestLine.RequestTarget == "/yourproblem" {
		handler400(w, req)
		return
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		handler500(w, req)
		return
	}
	if req.RequestLine.RequestTarget == "/video" {
		handlerVideo(w, req)
		return
	}
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		handlerHttpBin(w, req)
		return
	}
	handler200(w, req)
	return
}

func handlerHttpBin(w *response.Writer, req *request.Request) {
	h := response.GetDefaultHeaders(0)
	h.Remove("Content-Length")
	h.Override("Transfer-Encoding", "chunked")
	h["Trailer"] = "X-Content-SHA256, X-Content-Length"

	trailers := headers.NewHeaders()

	urlTarget := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	url := fmt.Sprintf("%s%s", "https://httpbin.org", urlTarget)
	log.Printf("The URL for the request is:  %v\n", url)
	extResponse, err := http.Get(url)
	var finalBody []byte
	if err != nil {
		body := fmt.Appendf([]byte{}, "Failed to get request from httpbin: %v", err)
		writeErrorBody(w, h, body)
		return
	}
	w.WriteStatusLine(response.StatusOK)
	w.WriteHeaders(h)
	const limit = 1024
	responseBuffer := make([]byte, limit, limit)
	for {
		n, err := extResponse.Body.Read(responseBuffer)
		if n > 0 {
			// Write whatever we got
			w.WriteChunkedBody(responseBuffer[:n])
			finalBody = append(finalBody, responseBuffer[:n]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			// Handle other errors
			break
		}
	}
	hash := sha256.Sum256(finalBody)
	trailers["X-Content-SHA256"] = fmt.Sprintf("%x", hash)
	trailers["X-Content-Length"] = strconv.Itoa(len(finalBody))
	w.WriteChunkedBodyDone(trailers)
	return
}

func handlerVideo(w *response.Writer, req *request.Request) {
	file, err := os.ReadFile("assets/vim.mp4")
	if err != nil {
		h := headers.NewHeaders()
		w.WriteStatusLine(response.StatusBadRequest)
		writeErrorBody(w, h, []byte(fmt.Sprintf("Failed to read file: %v", err)))
		return
	}
	length := strconv.Itoa(len(file))
	fmt.Printf("The length of the body is: %s\n", length)

	h := response.GetDefaultHeaders(len(file))
	h["Content-Type"] = "video/mp4"
	w.WriteStatusLine(response.StatusOK)
	err = w.WriteHeaders(h)
	if err != nil {
		log.Printf("Failed to write the headers to the buffer: %v\n", err)
	}
	w.WriteBody(file)
	return
}

func writeErrorBody(w *response.Writer, h headers.Headers, message []byte) {
	w.WriteStatusLine(response.StatusBadRequest)
	h["Content-Length"] = strconv.Itoa(len(message))
	w.WriteHeaders(h)
	w.WriteBody(message)
	return
}

func handler400(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusBadRequest)
	body := []byte(`<html>
<head>
<title>400 Bad Request</title>
</head>
<body>
<h1>Bad Request</h1>
<p>Your request honestly kinda sucked.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
	return
}

func handler500(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusInternalServerError)
	body := []byte(`<html>
<head>
<title>500 Internal Server Error</title>
</head>
<body>
<h1>Internal Server Error</h1>
<p>Okay, you know what? This one is on me.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler200(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusOK)
	body := []byte(`<html>
<head>
<title>200 OK</title>
</head>
<body>
<h1>Success!</h1>
<p>Your request was an absolute banger.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
	return
}

func writeResponse(statusCode response.StatusCode, writer *response.Writer, responseBody []byte) error {
	writer.WriterState = response.ReadyForStatusLine
	err := writer.WriteStatusLine(statusCode)
	if err != nil {
		return err
	}

	currHeaders := response.GetDefaultHeaders(len(responseBody))
	err = writer.WriteHeaders(currHeaders)
	if err != nil {
		log.Printf("Failed to write message to connection: %v\n", err)
		return err
	}
	_, err = writer.WriteBody(responseBody)
	if err != nil {
		log.Printf("Failed to send the response: %v\n", err)
		return err
	}
	return nil
}
