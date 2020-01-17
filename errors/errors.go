package errors

import "fmt"

func New(text string) error {
	return &errorString{fmt.Sprintf("gomig: %s", text)}
}

// errorString is a trivial implementation of error.
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}
