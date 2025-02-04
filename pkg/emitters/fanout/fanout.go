package fanout

import (
	"fmt"

	"github.com/appthrust/kutelog/pkg/core"
	"github.com/appthrust/kutelog/pkg/entry"
)

// Emitter broadcasts log entries to multiple emitters
type Emitter struct {
	emitters []core.Emitter
}

// NewEmitter creates a new fanout emitter with the given emitters
func NewEmitter(emitters ...core.Emitter) *Emitter {
	return &Emitter{
		emitters: emitters,
	}
}

// Init initializes all underlying emitters
func (e *Emitter) Init() error {
	for _, emitter := range e.emitters {
		if err := emitter.Init(); err != nil {
			return fmt.Errorf("failed to initialize emitter: %w", err)
		}
	}
	return nil
}

// Emit broadcasts the entry to all underlying emitters
func (e *Emitter) Emit(entry *entry.Entry) {
	for _, emitter := range e.emitters {
		emitter.Emit(entry)
	}
}
