package controller

import (
	"errors"
	"net/http"
)

type ErrorStatus struct {
	error
	statusCode int
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (e ErrorStatus) Unwrap() error { return e.error }

func ErrorInfo(err error) (ErrorResponse, int) {
	var errStatus ErrorStatus
	if errors.As(err, &errStatus) {
		return ErrorResponse{errStatus.error.Error()}, errStatus.statusCode
	}
	return ErrorResponse{errors.New("unknown error occurred").Error()}, http.StatusInternalServerError
}

func NewErrorStatus(err error, code int) error {
	return ErrorStatus{
		error:  err,
		statusCode: code,
	}
}