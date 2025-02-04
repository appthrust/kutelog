package logr_test

import (
	"io"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/appthrust/kutelog/pkg/entry"
	"github.com/appthrust/kutelog/pkg/parsers/logr"
)

var _ = Describe("Logr", func() {
	Describe("ParseLevel", func() {
		DescribeTable("parsing log levels",
			func(input string, expected entry.Level, expectError bool, errorContains string) {
				level, err := logr.ParseLevel(input)

				if expectError {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(errorContains))
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(level).To(Equal(expected))
				}
			},
			Entry("INFO level", "INFO", entry.LevelInfo, false, ""),
			Entry("WARNING level", "WARNING", entry.LevelWarning, false, ""),
			Entry("ERROR level", "ERROR", entry.LevelError, false, ""),
			Entry("DEBUG level", "DEBUG", entry.LevelDebug, false, ""),
			Entry("unknown level", "UNKNOWN", entry.Level(""), true, "unknown level: UNKNOWN"),
		)
	})

	Describe("IsStackNamish", func() {
		DescribeTable("identifying stack trace lines",
			func(input string, expected bool) {
				result := logr.IsStackNamish(input)
				Expect(result).To(Equal(expected))
			},
			Entry("valid stack trace line",
				"sigs.k8s.io/controller-runtime/pkg/internal/source.(*Kind[...]).Start.func1.1",
				true),
			Entry("valid package path",
				"github.com/example/pkg/file",
				true),
			Entry("RFC3339 timestamp line",
				"2025-01-30T15:52:37+09:00\tINFO\tsetup\tstarting manager",
				false),
			Entry("indented line",
				"        /path/to/file.go:123",
				false),
			Entry("tab indented line",
				"\t/path/to/file.go:123",
				false),
			Entry("plain text",
				"some random text",
				false),
		)
	})

	Describe("Parser", func() {
		var parser *logr.Parser

		BeforeEach(func() {
			parser = logr.NewParser()
		})

		Context("normal log line", func() {
			It("parses successfully", func() {
				input := "2025-01-30T15:52:37+09:00\tINFO\tsetup\tstarting manager\t{\"name\": \"test\"}"

				entries, err := parser.Parse(input, nil, nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(entries).To(HaveLen(1))

				e := entries[0]
				Expect(e.Structured.Timestamp.IsZero()).To(BeFalse())
				Expect(e.Structured.Level).To(Equal(entry.LevelInfo))
				Expect(e.Structured.Message).To(Equal("setup\tstarting manager"))
				Expect(e.Structured.Data).To(Equal(map[string]interface{}{
					"name": "test",
				}))
			})
		})

		Context("error log with stack trace", func() {
			It("parses log with stack trace successfully", func() {
				input := "2025-01-30T15:53:07+09:00\tERROR\tcontroller-runtime.source.EventHandler\tfailed to get informer from cache\t{\"error\": \"test error\"}"
				peekLines := []string{
					"sigs.k8s.io/controller-runtime/pkg/internal/source.(*Kind[...]).Start.func1.1",
					"        /Users/suin/go/pkg/mod/sigs.k8s.io/controller-runtime@v0.19.0/pkg/internal/source/kind.go:76",
					"2025-01-30T15:53:37+09:00", // next log line (end of stack trace)
				}

				var currentLine int
				peekLine := func() (string, error) {
					if currentLine >= len(peekLines) {
						return "", io.EOF
					}
					return peekLines[currentLine], nil
				}
				consumeLine := func() {
					currentLine++
				}

				entries, err := parser.Parse(input, peekLine, consumeLine)

				Expect(err).NotTo(HaveOccurred())
				Expect(entries).To(HaveLen(1))

				e := entries[0]
				Expect(e.Structured.Timestamp.IsZero()).To(BeFalse())
				Expect(e.Structured.Level).To(Equal(entry.LevelError))
				Expect(e.Structured.Message).To(Equal("controller-runtime.source.EventHandler\tfailed to get informer from cache"))
				Expect(e.Structured.Data).To(Equal(map[string]interface{}{
					"error": "test error",
				}))
				Expect(e.Structured.Stack).To(Equal(
					"sigs.k8s.io/controller-runtime/pkg/internal/source.(*Kind[...]).Start.func1.1\n" +
						"        /Users/suin/go/pkg/mod/sigs.k8s.io/controller-runtime@v0.19.0/pkg/internal/source/kind.go:76",
				))
			})
		})

		Context("log line without metadata (multiple spaces)", func() {
			It("parses successfully", func() {
				input := "2025-01-30T15:52:37+09:00\tINFO\tsetup\tstarting manager"
				entries, err := parser.Parse(input, nil, nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(entries).To(HaveLen(1))

				e := entries[0]
				Expect(e.Structured.Timestamp.IsZero()).To(BeFalse())
				Expect(e.Structured.Level).To(Equal(entry.LevelInfo))
				Expect(e.Structured.Message).To(Equal("setup\tstarting manager"))
				Expect(e.Structured.Data).To(Equal(map[string]interface{}{}))
			})
		})

		Context("log line without metadata (single space)", func() {
			It("parses successfully", func() {
				input := "2025-01-30T15:52:37+09:00\tINFO\tsetup\tstarting manager"
				entries, err := parser.Parse(input, nil, nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(entries).To(HaveLen(1))

				e := entries[0]
				Expect(e.Structured.Timestamp.IsZero()).To(BeFalse())
				Expect(e.Structured.Level).To(Equal(entry.LevelInfo))
				Expect(e.Structured.Message).To(Equal("setup\tstarting manager"))
				Expect(e.Structured.Data).To(Equal(map[string]interface{}{}))
			})
		})

		Context("error log with stack trace without metadata", func() {
			It("parses log with stack trace successfully", func() {
				input := "2025-01-30T15:53:07+09:00\tERROR\tcontroller-runtime.source.EventHandler\tfailed to get informer from cache"
				peekLines := []string{
					"sigs.k8s.io/controller-runtime/pkg/internal/source.(*Kind[...]).Start.func1.1",
					"        /Users/suin/go/pkg/mod/sigs.k8s.io/controller-runtime@v0.19.0/pkg/internal/source/kind.go:76",
					"2025-01-30T15:53:37+09:00", // next log line (end of stack trace)
				}

				var currentLine int
				peekLine := func() (string, error) {
					if currentLine >= len(peekLines) {
						return "", io.EOF
					}
					return peekLines[currentLine], nil
				}
				consumeLine := func() {
					currentLine++
				}

				entries, err := parser.Parse(input, peekLine, consumeLine)

				Expect(err).NotTo(HaveOccurred())
				Expect(entries).To(HaveLen(1))

				e := entries[0]
				Expect(e.Structured.Timestamp.IsZero()).To(BeFalse())
				Expect(e.Structured.Level).To(Equal(entry.LevelError))
				Expect(e.Structured.Message).To(Equal("controller-runtime.source.EventHandler\tfailed to get informer from cache"))
				Expect(e.Structured.Data).To(Equal(map[string]interface{}{}))
				Expect(e.Structured.Stack).To(Equal(
					"sigs.k8s.io/controller-runtime/pkg/internal/source.(*Kind[...]).Start.func1.1\n" +
						"        /Users/suin/go/pkg/mod/sigs.k8s.io/controller-runtime@v0.19.0/pkg/internal/source/kind.go:76",
				))
			})
		})

		Context("RFC3339 timestamp detection in stack trace", func() {
			It("should detect RFC3339 timestamp without timezone as non-stack-trace", func() {
				stream := `2025-02-01T00:00:00Z	ERROR	Something wrong	{"foo": "bar"}
example.com/path/to/file.(*Struct).Method
	/path/to/file.go:123
some.example.com/path/to/file.(*Struct).Method
	/path/to/file.go:456
other.example.com/path/to/file.(*Struct).Method
	/path/to/file.go:789
2025-02-01T00:00:00Z	INFO	Some message with package path-like string (e.g., sigs.k8s.io/)	{"baz": "qux"}`
				lines := strings.Split(stream, "\n")
				input := lines[0]
				peekLines := lines[1:]

				var currentLine int
				peekLine := func() (string, error) {
					if currentLine >= len(peekLines) {
						return "", io.EOF
					}
					return peekLines[currentLine], nil
				}
				consumeLine := func() {
					currentLine++
				}

				entries, err := parser.Parse(input, peekLine, consumeLine)

				Expect(err).NotTo(HaveOccurred())
				Expect(entries).To(HaveLen(1))

				e := entries[0]
				Expect(e.Structured.Level).To(Equal(entry.LevelError))
				Expect(e.Structured.Message).To(Equal("Something wrong"))
				Expect(e.Structured.Data).To(Equal(map[string]interface{}{
					"foo": "bar",
				}))

				expectedStack := `example.com/path/to/file.(*Struct).Method
	/path/to/file.go:123
some.example.com/path/to/file.(*Struct).Method
	/path/to/file.go:456
other.example.com/path/to/file.(*Struct).Method
	/path/to/file.go:789`
				Expect(e.Structured.Stack).To(Equal(expectedStack))
			})
		})

		Context("error cases", func() {
			It("returns error when metadata is invalid JSON", func() {
				input := "2025-01-30T15:52:37+09:00\tINFO\tsetup\tstarting manager\t{\"invalid\": json}"
				_, err := parser.Parse(input, nil, nil)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to unmarshal metadata"))
			})
		})
	})
})
