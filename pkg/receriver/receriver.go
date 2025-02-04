package receriver

import (
	"bufio"
	"io"

	"github.com/appthrust/kutelog/pkg/core"
	"github.com/appthrust/kutelog/pkg/entry"
)

var _ core.Receiver = &Receiver{}

type Receiver struct {
	parser Parser
}

func NewReceiver(parser Parser) *Receiver {
	return &Receiver{parser: parser}
}

func (r *Receiver) Receive(input io.Reader, entriesChan chan<- *entry.Entry, errChan chan<- error) {
	scanner := bufio.NewScanner(input)

	var currentLine string // current line being processed
	var nextLine string    // next line
	var hasPeeked bool     // whether peek has been performed

	// internal function to read next line
	readNextLine := func() (string, error) {
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return "", err
			}
			return "", io.EOF
		}
		return scanner.Text(), nil
	}

	peekLine := func() (string, error) {
		if !hasPeeked {
			// read new line if peek hasn't been performed yet
			line, err := readNextLine()
			if err != nil {
				return "", err
			}
			nextLine = line
			hasPeeked = true
		}
		return nextLine, nil
	}

	consumeLine := func() {
		hasPeeked = false // reset for next peek to read new line
	}

	// main loop
	for {
		// read line
		line, err := readNextLine()
		if err != nil {
			if err != io.EOF {
				errChan <- err
				return
			}
			// wait for new input on EOF
			continue
		}
		currentLine = line

		entries, err := r.parser.Parse(currentLine, peekLine, consumeLine)
		if err != nil {
			errChan <- err
			return
		}

		for _, entry := range entries {
			entriesChan <- entry
		}

		// get next line (if peeked)
		if hasPeeked {
			// use the peeked but not consumed line from previous Parser as next currentLine
			currentLine = nextLine
			hasPeeked = false // reset state for new Parser
		}
	}
}

// Parser is an interface that parses one or more lines of text to generate entries
type Parser interface {
	// Parse parses the given line and returns a slice of entries
	// peekLine: function to peek at next line (without consuming it)
	// consumeLine: function to consume the next line (advance)
	Parse(line string, peekLine func() (string, error), consumeLine func()) ([]*entry.Entry, error)
}
