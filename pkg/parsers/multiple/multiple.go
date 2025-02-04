package multiple

import (
	"github.com/appthrust/kutelog/pkg/entry"
	"github.com/appthrust/kutelog/pkg/receriver"
)

var _ receriver.Parser = &Parser{}

type Parser struct {
	parsers []receriver.Parser
}

func NewParser(parsers ...receriver.Parser) *Parser {
	return &Parser{parsers: parsers}
}

func (p *Parser) Parse(line string, peekLine func() (string, error), consumeLine func()) ([]*entry.Entry, error) {
	for _, parser := range p.parsers {
		entries, err := parser.Parse(line, peekLine, consumeLine)
		if err == nil {
			return entries, nil
		}
	}
	return []*entry.Entry{{Unstructured: line}}, nil
}
