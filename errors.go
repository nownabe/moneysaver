package main

import (
	"fmt"
	"net/http"
)

type httpError struct {
	status int
	err    error
}

func e(status int, msg string, args ...any) error {
	return httpError{
		status: status,
		err:    fmt.Errorf(msg, args...),
	}
}

func (e httpError) Error() string {
	return e.err.Error()
}

func (e httpError) Status() int {
	return e.status
}

func writeErrorHeader(w http.ResponseWriter, err error) {
	e, ok := err.(httpError)
	if ok {
		w.WriteHeader(e.Status())
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
