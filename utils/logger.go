package utils

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// Logger is a minimal thread-safe logger for the installer.
type Logger struct {
	mu     sync.Mutex
	writer io.Writer
}

// NewLogger creates a logger writing to stdout.
func NewLogger() *Logger {
	return &Logger{writer: os.Stdout}
}

// Log emits a scoped installer line.
func (l *Logger) Log(scope, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.writer, "[%s] %s\n", scope, message)
}
