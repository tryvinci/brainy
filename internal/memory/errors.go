package memory

import "errors"

var (
	ErrMemoryNotFound = errors.New("memory not found")
	ErrMemoryConflict = errors.New("memory conflict")
	// ErrLeaseLost is returned when a worker tries to complete or fail a job
	// whose lease no longer belongs to it (expired and reclaimed by another
	// worker). The caller must abort without writing results.
	ErrLeaseLost = errors.New("job lease lost to another worker")
)
