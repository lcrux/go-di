package utils

import "testing"

func TestShutdownHandler_ReturnsWaitGroup(t *testing.T) {
	wg := ShutdownHandler(func() {})
	if wg == nil {
		t.Fatal("expected non-nil wait group")
	}
}
