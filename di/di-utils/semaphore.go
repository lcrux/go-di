package diutils

import (
	"os"
	"strconv"
)

var defaultSemaphoreCapacity int = 10

func init() {
	if envLimit := os.Getenv("GODI_SEMAPHORE_CAPACITY"); envLimit != "" {
		if val, err := strconv.Atoi(envLimit); err == nil && val > 0 {
			defaultSemaphoreCapacity = val
		}
	}
}

// Semaphore is a simple semaphore implementation.
type Semaphore struct {
	ch chan struct{}
}

// NewSemaphore creates a new semaphore with the given capacity.
// If the capacity is less than or equal to 0, it defaults to 10.
func NewSemaphore(capacity ...int) *Semaphore {
	if len(capacity) == 0 || capacity[0] <= 0 {
		capacity = []int{defaultSemaphoreCapacity}
	}
	return &Semaphore{
		ch: make(chan struct{}, capacity[0]),
	}
}

// Acquire acquires a slot in the semaphore, blocking if necessary.
func (s *Semaphore) Acquire() {
	s.ch <- struct{}{}
}

// Release releases a slot in the semaphore.
func (s *Semaphore) Release() {
	<-s.ch
}

// Done closes the semaphore channel, releasing all resources.
// Any attempt to acquire or release the semaphore after calling Done will panic.
func (s *Semaphore) Done() {
	close(s.ch)
}
