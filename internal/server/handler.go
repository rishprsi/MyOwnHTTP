package server

import (
	"MyOwnHTTP/internal/request"
	"MyOwnHTTP/internal/response"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w *response.Writer, req *request.Request)
