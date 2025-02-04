package stdout_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/appthrust/kutelog/pkg/emitters/stdout"
	"github.com/appthrust/kutelog/pkg/entry"
)

var _ = Describe("Stdout Emitter", func() {
	var (
		emitter   *stdout.Emitter
		output    *bytes.Buffer
		oldStdout *os.File
		w         *os.File
		wg        sync.WaitGroup
	)

	BeforeEach(func() {
		output = new(bytes.Buffer)
		oldStdout = os.Stdout
		r, wr, err := os.Pipe()
		Expect(err).NotTo(HaveOccurred())
		w = wr
		os.Stdout = w

		wg.Add(1)
		go func() {
			defer wg.Done()
			io.Copy(output, r)
		}()

		emitter = stdout.NewEmitter()
		Expect(emitter.Init()).To(Succeed())
	})

	AfterEach(func() {
		w.Close()
		wg.Wait()
		os.Stdout = oldStdout
	})

	Context("when emitting entries", func() {
		It("writes structured logs as JSON", func() {
			timestamp := time.Now()
			testEntry := &entry.Entry{
				Structured: &entry.Structured{
					Timestamp: timestamp,
					Level:     entry.LevelInfo,
					Message:   "test message",
					Data: map[string]interface{}{
						"key": "value",
					},
				},
			}
			emitter.Emit(testEntry)
			w.Close()
			wg.Wait()

			var received entry.Structured
			err := json.NewDecoder(output).Decode(&received)
			Expect(err).NotTo(HaveOccurred())
			Expect(received.Timestamp).To(BeTemporally("~", timestamp, time.Second))
			Expect(received.Level).To(Equal(entry.LevelInfo))
			Expect(received.Message).To(Equal("test message"))
			Expect(received.Data).To(HaveKeyWithValue("key", "value"))
		})

		It("writes unstructured logs as plain text", func() {
			testEntry := &entry.Entry{
				Unstructured: "plain text log",
			}
			emitter.Emit(testEntry)
			w.Close()
			wg.Wait()

			Expect(output.String()).To(Equal("plain text log\n"))
		})

		It("handles empty entries", func() {
			testEntry := &entry.Entry{}
			emitter.Emit(testEntry)
			w.Close()
			wg.Wait()

			Expect(output.String()).To(BeEmpty())
		})
	})
})
