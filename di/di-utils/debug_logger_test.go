package diutils

import (
	"log"
	"os"
	"strings"
	"testing"
)

func TestDebugLog(t *testing.T) {
	os.Setenv("GODI_DEBUG", "true")
	defer os.Unsetenv("GODI_DEBUG")

	// Redirect log output for testing
	logOutput := &mockWriter{}
	log.SetOutput(logOutput)       // Redirect log output to the mock writer
	defer log.SetOutput(os.Stderr) // Restore the original log output after the test

	log.SetFlags(0)
	defer log.SetFlags(log.LstdFlags)

	DebugLog("Test message: %d", 42)

	if strings.Trim(logOutput.lastMessage, "\n") == "" {
		t.Fatal("Expected log message to be written")
	}

	if strings.Trim(logOutput.lastMessage, "\n") != "Test message: 42" {
		t.Fatalf("Expected 'Test message: 42', got '%s'", logOutput.lastMessage)
	}
}

func TestDebugLog_Disabled(t *testing.T) {
	os.Unsetenv("GODI_DEBUG")

	logOutput := &mockWriter{}
	log.SetOutput(logOutput)
	defer log.SetOutput(os.Stderr)

	log.SetFlags(0)
	defer log.SetFlags(log.LstdFlags)

	DebugLog("Should not log")

	if strings.Trim(logOutput.lastMessage, "\n") != "" {
		t.Fatalf("Expected no log output when disabled, got '%s'", logOutput.lastMessage)
	}
}

type mockWriter struct {
	lastMessage string
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	m.lastMessage = string(p)
	return len(p), nil
}
