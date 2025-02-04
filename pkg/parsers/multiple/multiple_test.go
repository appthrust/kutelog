package multiple_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/appthrust/kutelog/pkg/entry"
	"github.com/appthrust/kutelog/pkg/parsers/multiple"
	"github.com/appthrust/kutelog/pkg/receriver"
)

var _ receriver.Parser = &mockParser{}

type mockParser struct {
	shouldError bool
	entries     []*entry.Entry
}

func (p *mockParser) Parse(line string, peekLine func() (string, error), consumeLine func()) ([]*entry.Entry, error) {
	if p.shouldError {
		return nil, &mockError{msg: "mock error"}
	}
	return p.entries, nil
}

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

var _ = Describe("Multiple", func() {
	Describe("Parser", func() {
		Context("trying multiple parsers in sequence", func() {
			It("succeeds with first parser", func() {
				parsers := []receriver.Parser{
					&mockParser{
						shouldError: false,
						entries: []*entry.Entry{{
							Structured: &entry.Structured{
								Message: "first parser",
							},
						}},
					},
					&mockParser{
						shouldError: false,
						entries: []*entry.Entry{{
							Structured: &entry.Structured{
								Message: "second parser",
							},
						}},
					},
				}

				parser := multiple.NewParser(parsers...)
				entries, err := parser.Parse("test line", nil, nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(entries).To(HaveLen(1))
				Expect(entries[0].Structured.Message).To(Equal("first parser"))
			})

			It("succeeds with second parser when first fails", func() {
				parsers := []receriver.Parser{
					&mockParser{
						shouldError: true,
					},
					&mockParser{
						shouldError: false,
						entries: []*entry.Entry{{
							Structured: &entry.Structured{
								Message: "second parser",
							},
						}},
					},
				}

				parser := multiple.NewParser(parsers...)
				entries, err := parser.Parse("test line", nil, nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(entries).To(HaveLen(1))
				Expect(entries[0].Structured.Message).To(Equal("second parser"))
			})

			It("falls back to unstructured when all parsers fail", func() {
				parsers := []receriver.Parser{
					&mockParser{shouldError: true},
					&mockParser{shouldError: true},
				}

				parser := multiple.NewParser(parsers...)
				entries, err := parser.Parse("test line", nil, nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(entries).To(HaveLen(1))
				Expect(entries[0].Unstructured).To(Equal("test line"))
			})
		})
	})
})
