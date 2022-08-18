package main

import (
	"errors"
	"fmt"
	"net/http"
)

type httpError struct {
	status int
	err    error
}

func e(status int, msg string) error {
	return httpError{
		status: status,
		err:    errors.New(msg),
	}
}

func wrap(status int, msg string, e error) error {
	return httpError{
		status: status,
		err:    fmt.Errorf(msg, e),
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
