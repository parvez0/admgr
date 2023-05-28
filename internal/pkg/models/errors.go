package models

import (
	"fmt"
)

const (
	DecodeFailureError             = 1
	InternalProcessingError        = 2
	DuplicateResourceCreationError = 3
	ResourceNotFoundError          = 4
	ActionForbidden                = 5
	DetailedResourceInfoNotFound   = 6
	DependentServiceRequestFailed  = 7
)

type Error struct {
	Type    int
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("Code: %d, Error: %s", e.Type, e.Message)
}

func NewError(message string, code int) error {
	return &Error{
		Type:    code,
		Message: message,
	}
}
