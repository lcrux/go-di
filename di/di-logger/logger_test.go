package dilogger

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestInitLogLevel_Info(t *testing.T) {
	initLogLevel("info")
	if defaultLogLevel != Info {
		t.Fatalf("expected Info, got %d", defaultLogLevel)
	}
}

func TestInitLogLevel_Debug(t *testing.T) {
	initLogLevel("debug")
	if defaultLogLevel != Debug {
		t.Fatalf("expected Debug, got %d", defaultLogLevel)
	}
}

func TestInitLogLevel_Warn(t *testing.T) {
	initLogLevel("WARN")
	if defaultLogLevel != Warn {
		t.Fatalf("expected Warn, got %d", defaultLogLevel)
	}
}

func TestInitLogLevel_Error(t *testing.T) {
	initLogLevel("ERROR")
	if defaultLogLevel != Error {
		t.Fatalf("expected Error, got %d", defaultLogLevel)
	}
}

func TestInitLogLevel_Empty_DefaultsToError(t *testing.T) {
	initLogLevel("")
	if defaultLogLevel != Error {
		t.Fatalf("expected Error for empty string, got %d", defaultLogLevel)
	}
}

func TestInitLogLevel_Unknown_DefaultsToErrorAndLogs(t *testing.T) {
	initLogLevel("INVALID")
	if defaultLogLevel != Error {
		t.Fatalf("expected Error for unknown level, got %d", defaultLogLevel)
	}
}

func TestLoggerImpl_LogLevelFiltering(t *testing.T) {
	called := make(map[string]bool)
	opts := &LoggerOptions{
		LogLevel: Warn,
		Info:     func(_ string, _ ...interface{}) { called["info"] = true },
		Warn:     func(_ string, _ ...interface{}) { called["warn"] = true },
		Debug:    func(_ string, _ ...interface{}) { called["debug"] = true },
		Error:    func(_ string, _ ...interface{}) { called["error"] = true },
	}
	logger := NewLogger(func(o *LoggerOptions) {
		*o = *opts
	})
	logger.Infof("Should not log info")
	logger.Debugf("Should not log debug")
	logger.Warnf("Should log warn")
	logger.Errorf("Should log error")
	if called["info"] {
		t.Errorf("Info should not be called at Warn level")
	}
	if called["debug"] {
		t.Errorf("Debug should not be called at Warn level")
	}
	if !called["warn"] {
		t.Errorf("Warn should be called at Warn level")
	}
	if !called["error"] {
		t.Errorf("Error should be called at Warn level")
	}
}

func TestLoggerImpl_CustomLoggerFuncs(t *testing.T) {
	var messages []string
	opts := &LoggerOptions{
		LogLevel: Info,
		Info:     func(format string, v ...interface{}) { messages = append(messages, fmt.Sprintf(format, v...)) },
		Warn:     func(format string, v ...interface{}) { messages = append(messages, fmt.Sprintf(format, v...)) },
		Debug:    func(format string, v ...interface{}) { messages = append(messages, fmt.Sprintf(format, v...)) },
		Error:    func(format string, v ...interface{}) { messages = append(messages, fmt.Sprintf(format, v...)) },
	}
	logger := NewLogger(func(o *LoggerOptions) {
		*o = *opts
	})
	logger.Infof("Info message")
	logger.Warnf("Warn message")
	logger.Debugf("Debug message")
	logger.Errorf("Error message")
	expected := []string{"Info message", "Warn message", "Debug message", "Error message"}
	for i, msg := range expected {
		if messages[i] != msg {
			t.Fatalf("Expected '%s', got '%s'", msg, messages[i])
		}
	}
}

func TestLoggerImpl_EmptyMessage(t *testing.T) {
	opts := &LoggerOptions{
		LogLevel: Info,
		Info: func(format string, _ ...interface{}) {
			if format != "" {
				t.Fatalf("Expected empty message, got '%s'", format)
			}
		},
	}
	logger := NewLogger(func(o *LoggerOptions) {
		*o = *opts
	})
	logger.Infof("")
}

func TestLoggerImpl_FormatString(t *testing.T) {
	var got string
	opts := &LoggerOptions{
		LogLevel: Info,
		Info:     func(format string, v ...interface{}) { got = fmt.Sprintf(format, v...) },
	}
	logger := NewLogger(func(o *LoggerOptions) {
		*o = *opts
	})
	logger.Infof("Formatted message: %d, %s", 42, "test")
	expected := "Formatted message: 42, test"
	if got != expected {
		t.Fatalf("Expected '%s', got '%s'", expected, got)
	}
}

func TestLoggerImpl_ConcurrentLogging(t *testing.T) {
	var mu sync.Mutex
	var messages []string
	opts := &LoggerOptions{
		LogLevel: Info,
		Info: func(format string, v ...interface{}) {
			mu.Lock()
			messages = append(messages, fmt.Sprintf(format, v...))
			mu.Unlock()
		},
	}
	logger := NewLogger(func(o *LoggerOptions) {
		*o = *opts
	})
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		logger.Infof("Message from goroutine 1")
	}()
	go func() {
		defer wg.Done()
		logger.Infof("Message from goroutine 2")
	}()
	wg.Wait()
	// Ensure no panic occurred; exact message order is not guaranteed
	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}
}

func TestLoggerImpl_LongMessage(t *testing.T) {
	var got string
	opts := &LoggerOptions{
		LogLevel: Info,
		Info:     func(format string, v ...interface{}) { got = fmt.Sprintf(format, v...) },
	}
	logger := NewLogger(func(o *LoggerOptions) {
		*o = *opts
	})
	longMessage := fmt.Sprintf("This is a very long message. %v", strings.Repeat("A", 1000))
	logger.Infof("%s", longMessage)
	if got != longMessage {
		t.Fatalf("Expected long message, got '%s'", got)
	}
}

func TestLoggerImpl_SpecialCharacters(t *testing.T) {
	var got string
	opts := &LoggerOptions{
		LogLevel: Info,
		Info:     func(format string, v ...interface{}) { got = fmt.Sprintf(format, v...) },
	}
	logger := NewLogger(func(o *LoggerOptions) {
		*o = *opts
	})
	message := "Message with special characters: \n\t\u2713"
	logger.Infof("%s", message)
	if got != message {
		t.Fatalf("Expected '%s', got '%s'", message, got)
	}
}
