package actions

import (
	"fmt"
)

type StandardError struct {
	Message string
}

func (e *StandardError) Error() string {
	return e.Message
}

func ErrF(format string, parameters ...interface{}) error {
	return &StandardError{Message: fmt.Sprintf(format, parameters...)}
}
