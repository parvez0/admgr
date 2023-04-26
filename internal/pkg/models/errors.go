package models

const (
	DecodeFailureError             = 1
	InternalProcessingError        = 2
	DuplicateResourceCreationError = 3
	ResourceNotFoundError          = 4
	ActionForbidden                = 5
	DetailedResourceInfoNotFound   = 6
)

type Error struct {
	Type    int
	Message string
}
