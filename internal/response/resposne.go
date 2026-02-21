package response

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"

	"MyOwnHTTP/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

type WriterState int

const (
	ReadyForStatusLine WriterState = iota
	ReadyForHeader
	ReadyForBody
	Done
)

type Writer struct {
	Buffer      io.Writer
	WriterState WriterState
}

func NewWriter(conn net.Conn) *Writer {
	return &Writer{
		Buffer:      conn,
		WriterState: ReadyForStatusLine,
	}
}

func GetDefaultStatusLine(w io.Writer, statusCode StatusCode) error {
	var err error
	switch statusCode {
	case StatusOK:
		_, err = w.Write([]byte("HTTP/1.1 200 OK\r\n"))
	case StatusBadRequest:
		_, err = w.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
	case StatusInternalServerError:
		_, err = w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
	default:
		_, err = w.Write([]byte("HTTP/1.1 200"))
		log.Printf("Invalid status code recieved: %d\n", statusCode)
	}
	if err != nil {
		log.Printf("Failed to write status code: %v\n", err)
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	currHeaders := headers.NewHeaders()
	currHeaders["Content-length"] = strconv.Itoa(contentLen)
	currHeaders["Connection"] = "close"
	currHeaders["Content-Type"] = "text/html"

	return currHeaders
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.WriterState != ReadyForStatusLine {
		return fmt.Errorf("incorrect order of operations expected: %v, Got: %v", ReadyForStatusLine, w.WriterState)
	}
	err := GetDefaultStatusLine(w.Buffer, statusCode)
	if err != nil {
		return nil
	}
	w.WriterState = ReadyForHeader
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.WriterState != ReadyForHeader {
		return fmt.Errorf("incorrect order of operations expected: %v, Got: %v", ReadyForHeader, w.WriterState)
	}
	message := ""
	for key, value := range headers {
		message = fmt.Sprintf("%s%s:%s\r\n", message, key, value)
	}
	message += "\r\n"
	_, err := w.Buffer.Write([]byte(message))
	if err != nil {
		return err
	}
	w.WriterState = ReadyForBody
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.WriterState != ReadyForBody {
		return 0, fmt.Errorf("incorrect order of operations expected: %v, Got: %v", ReadyForBody, w.WriterState)
	}
	w.WriterState = Done
	return w.Buffer.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.WriterState != ReadyForBody {
		return 0, fmt.Errorf("incorrect order of operations expected: %v, Got: %v", ReadyForBody, w.WriterState)
	}
	hexLength := fmt.Sprintf("%x", len(p))
	responseBody := fmt.Sprintf("%s\r\n%s\r\n", hexLength, string(p))
	fmt.Println("The response body is: ", responseBody)
	n, err := w.Buffer.Write([]byte(responseBody))
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (w *Writer) WriteChunkedBodyDone(trailers headers.Headers) (int, error) {
	message := []byte("0\r\n")
	_, err := w.Buffer.Write(message)
	if err != nil {
		return 0, err
	}
	err = w.WriteTrailers(trailers)
	_, err = w.Buffer.Write([]byte("\r\n"))
	if err != nil {
		return 0, err
	}
	w.WriterState = Done
	return 0, nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	message := ""
	for key, value := range h {
		message = fmt.Sprintf("%s%s:%s\r\n", message, key, value)
	}
	message += "\r\n"
	_, err := w.Buffer.Write([]byte(message))
	if err != nil {
		return err
	}
	w.WriterState = ReadyForBody
	return nil
}
