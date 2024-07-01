package errors

// NotFoundError represents a resource not found error
type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

// DockerErrors represents errors coming from the docker service
type DockerError struct {
	Message string
}

func (e *DockerError) Error() string {
	return e.Message
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// InternalError represents an internal server error
type InternalError struct {
	Message string
}

func (e *InternalError) Error() string {
	return e.Message
}

func NewNotFoundError(message string) error {
	return &NotFoundError{Message: message}
}

func NewValidationError(message string) error {
	return &ValidationError{Message: message}
}

func NewInternalError(message string) error {
	return &InternalError{Message: message}
}

func NewDockerError(message string) error {
	return &InternalError{Message: message}
}
