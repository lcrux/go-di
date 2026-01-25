package libutils

import (
	"sync"
	"testing"
	"time"
)

func TestNewSemaphore(t *testing.T) {
	// Test with valid capacity
	sem := NewSemaphore(3)
	if cap(sem.ch) != 3 {
		t.Fatalf("Expected capacity 3, got %d", cap(sem.ch))
	}

	// Test with invalid capacity (<= 0)
	sem = NewSemaphore(0)
	if cap(sem.ch) != defaultSemaphoreCapacity {
		t.Fatalf("Expected default capacity %d, got %d", defaultSemaphoreCapacity, cap(sem.ch))
	}
}

func TestSemaphoreAcquireRelease(t *testing.T) {
	sem := NewSemaphore(2)

	// Acquire twice (should not block)
	sem.Acquire()
	sem.Acquire()

	// Use a goroutine to test blocking behavior
	acquired := false
	go func() {
		sem.Acquire()
		acquired = true
	}()

	// Wait a bit to ensure the goroutine is blocked
	time.Sleep(100 * time.Millisecond)
	if acquired {
		t.Fatal("Expected Acquire to block, but it did not")
	}

	// Release one slot and ensure the goroutine proceeds
	sem.Release()
	time.Sleep(100 * time.Millisecond)
	if !acquired {
		t.Fatal("Expected Acquire to proceed after Release, but it did not")
	}
}

func TestSemaphoreDone(t *testing.T) {
	sem := NewSemaphore(2)

	// Acquire and release to ensure semaphore is functional
	sem.Acquire()
	sem.Release()

	// Call Done and ensure the channel is closed
	sem.Done()

	// Further operations on the semaphore should panic
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic after Done, but no panic occurred")
		}
	}()
	sem.Acquire()
}

func TestSemaphoreConcurrency(t *testing.T) {
	sem := NewSemaphore(3)
	var wg sync.WaitGroup
	var counter int
	mu := sync.Mutex{}

	// Start 10 goroutines that acquire and release the semaphore
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem.Acquire()
			mu.Lock()
			counter++
			mu.Unlock()
			time.Sleep(50 * time.Millisecond) // Simulate work
			sem.Release()
		}()
	}

	wg.Wait()

	// Ensure the counter was incremented 10 times
	if counter != 10 {
		t.Fatalf("Expected counter to be 10, got %d", counter)
	}
}
