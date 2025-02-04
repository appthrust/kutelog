package logr

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/appthrust/kutelog/pkg/entry"
	"github.com/appthrust/kutelog/pkg/receriver"
)

var rfc3339Regex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`)

var _ receriver.Parser = &Parser{}

type Parser struct {
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(line string, peekLine func() (string, error), consumeLine func()) ([]*entry.Entry, error) {
	// Process JSON metadata if exists
	var data map[string]interface{}
	textPart := line
	if idx := strings.Index(line, "{"); idx != -1 {
		jsonPart := line[idx:]
		if err := json.Unmarshal([]byte(jsonPart), &data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		textPart = line[:idx-1]
	} else {
		data = make(map[string]interface{})
	}

	textPartTokens := strings.Split(textPart, "\t")
	if len(textPartTokens) < 3 {
		return nil, fmt.Errorf("invalid log format: missing tab")
	}

	level, err := ParseLevel(textPartTokens[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse level: %w", err)
	}

	timestamp, err := time.Parse(time.RFC3339, textPartTokens[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	message := strings.Join(textPartTokens[2:], "\t")

	// collect stack trace
	var stack string
	if level == entry.LevelError {
		stack = parseStack(peekLine, consumeLine)
	}

	return []*entry.Entry{{
		Structured: &entry.Structured{
			Timestamp: timestamp,
			Level:     level,
			Message:   message,
			Data:      data,
			Stack:     stack,
		},
	}}, nil
}

// ParseLevel converts log level string to entry.Level
func ParseLevel(level string) (entry.Level, error) {
	switch level {
	case "INFO":
		return entry.LevelInfo, nil
	case "WARNING":
		return entry.LevelWarning, nil
	case "ERROR":
		return entry.LevelError, nil
	case "DEBUG":
		return entry.LevelDebug, nil
	default:
		return "", fmt.Errorf("unknown level: %s", level)
	}
}

// IsStackNamish determines if a line is a function name line in stack trace
func IsStackNamish(line string) bool {
	// no indentation
	if strings.HasPrefix(line, "\t") || strings.HasPrefix(line, " ") {
		return false
	}
	// verify it's not a date (RFC3339)
	if strings.HasPrefix(line, "2") { // starts with 2XXX
		// check for date-like pattern
		if line[4] == '-' && line[7] == '-' && line[10] == 'T' {
			// verify if it matches RFC3339 format
			if rfc3339Regex.MatchString(line) {
				return false
			}
		}
	}
	// starts with package path-like string (e.g., sigs.k8s.io/)
	return strings.Contains(line, "/") && strings.Contains(line, ".")
}

// isStackPositionish determines if a line is a position info line in stack trace
func isStackPositionish(line string) bool {
	// indented with tab or 8 spaces
	if !strings.HasPrefix(line, "\t") && !strings.HasPrefix(line, "        ") {
		return false
	}
	// contains '.go:'
	if !strings.Contains(line, ".go:") {
		return false
	}
	// ends with line number
	lastChar := line[len(line)-1]
	return '0' <= lastChar && lastChar <= '9'
}

func parseStack(peekLine func() (string, error), consumeLine func()) string {
	var stackLines []string
	expect := "stack_name"

	for {
		next, err := peekLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "" // return empty string on error
		}
		if expect == "stack_name" {
			if !IsStackNamish(next) {
				break
			}
			expect = "stack_position"
		} else if expect == "stack_position" {
			if !isStackPositionish(next) {
				break
			}
			expect = "stack_name"
		}
		// add stack trace line
		stackLines = append(stackLines, next)
		consumeLine() // consume the line
	}

	if len(stackLines) == 0 {
		return ""
	}

	return strings.Join(stackLines, "\n")
}
