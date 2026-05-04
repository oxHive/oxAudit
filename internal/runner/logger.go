package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/graditya/oxaudit/internal/types"
)

// CommandLogger appends structured JSON lines to commands.jsonl and errors.jsonl.
type CommandLogger struct {
	mu       sync.Mutex
	cmdFile  *os.File
	errFile  *os.File
}

// NewCommandLogger opens the log files for appending.
func NewCommandLogger(logsDir string) (*CommandLogger, error) {
	cmdPath := fmt.Sprintf("%s/commands.jsonl", logsDir)
	errPath := fmt.Sprintf("%s/errors.jsonl", logsDir)

	cmdFile, err := os.OpenFile(cmdPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("opening commands log: %w", err)
	}
	errFile, err := os.OpenFile(errPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		cmdFile.Close()
		return nil, fmt.Errorf("opening errors log: %w", err)
	}
	return &CommandLogger{cmdFile: cmdFile, errFile: errFile}, nil
}

// Log writes a RawFile log entry. Safe for concurrent use.
func (l *CommandLogger) Log(rf types.RawFile) {
	l.mu.Lock()
	defer l.mu.Unlock()

	data, _ := json.Marshal(rf)
	data = append(data, '\n')

	l.cmdFile.Write(data)
	if rf.Status != "ok" {
		l.errFile.Write(data)
	}
}

// Close flushes and closes both log files.
func (l *CommandLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	e1 := l.cmdFile.Close()
	e2 := l.errFile.Close()
	if e1 != nil {
		return e1
	}
	return e2
}
