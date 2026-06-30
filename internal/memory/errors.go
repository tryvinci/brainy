package memory

import "errors"

var (
	ErrMemoryNotFound = errors.New("memory not found")
	ErrMemoryConflict = errors.New("memory conflict")
)
