package stdout

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/appthrust/kutelog/pkg/entry"
)

// Emitter writes log entries to stdout
type Emitter struct {
	encoder *json.Encoder
}

// NewEmitter creates a new stdout emitter
func NewEmitter() *Emitter {
	return &Emitter{
		encoder: json.NewEncoder(os.Stdout),
	}
}

// Init initializes the emitter
func (e *Emitter) Init() error {
	return nil
}

// Emit writes the entry to stdout
func (e *Emitter) Emit(entry *entry.Entry) {
	if entry.Structured != nil {
		e.encoder.Encode(entry.Structured)
	} else if entry.Unstructured != "" {
		fmt.Fprintln(os.Stdout, entry.Unstructured)
	}
}
