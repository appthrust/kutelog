package entry

import "time"

type Entry struct {
	Structured   *Structured
	Unstructured string
}

type Structured struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     Level                  `json:"level"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Stack     string                 `json:"stack,omitempty"`
}

type Level string

const (
	LevelInfo    Level = "info"
	LevelWarning Level = "warning"
	LevelError   Level = "error"
	LevelDebug   Level = "debug"
)
