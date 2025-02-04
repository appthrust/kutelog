package core

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/appthrust/kutelog/pkg/entry"
)

type Process struct {
	receiver Receiver
	emitter  Emitter
}

func NewProcess(options *ProcessOptions) *Process {
	return &Process{
		receiver: options.Receiver,
		emitter:  options.Emitter,
	}
}

func (p *Process) Start() error {
	entries := make(chan *entry.Entry, 1000)
	errChan := make(chan error)
	// Must start receiver before emitter since emitter initialization may be slow
	go p.receiver.Receive(os.Stdin, entries, errChan)
	if err := p.emitter.Init(); err != nil {
		return fmt.Errorf("failed to initialize emitter: %w", err)
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case e := <-entries:
			p.emitter.Emit(e)
		case err := <-errChan:
			return fmt.Errorf("receiver error: %w", err)
		case <-sig:
			return nil
		}
	}
}

type ProcessOptions struct {
	Receiver Receiver
	Emitter  Emitter
}

type Receiver interface {
	Receive(input io.Reader, entries chan<- *entry.Entry, err chan<- error)
}

type Emitter interface {
	Init() error
	Emit(*entry.Entry)
}
