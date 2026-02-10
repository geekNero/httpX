package server

import (
	"basic_protocol/internal/request"
	"basic_protocol/internal/response"
)

type HandlerError struct {
	StatusCode response.StatusCode `json:"status_code"`
	Message    string              `json:"message"`
}

type Handler func(w *response.Writer, req *request.Request)
