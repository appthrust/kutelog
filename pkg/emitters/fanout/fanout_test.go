package fanout_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/appthrust/kutelog/pkg/emitters/fanout"
	"github.com/appthrust/kutelog/pkg/entry"
)

// mockEmitter is a test double that implements the core.Emitter interface
type mockEmitter struct {
	initErr    error
	initCalled bool
	entries    []*entry.Entry
}

func (m *mockEmitter) Init() error {
	m.initCalled = true
	return m.initErr
}

func (m *mockEmitter) Emit(e *entry.Entry) {
	m.entries = append(m.entries, e)
}

var _ = Describe("Fanout Emitter", func() {
	var (
		mock1   *mockEmitter
		mock2   *mockEmitter
		emitter *fanout.Emitter
	)

	BeforeEach(func() {
		mock1 = &mockEmitter{}
		mock2 = &mockEmitter{}
		emitter = fanout.NewEmitter(mock1, mock2)
	})

	Context("when initializing", func() {
		It("initializes all emitters successfully", func() {
			Expect(emitter.Init()).To(Succeed())
			Expect(mock1.initCalled).To(BeTrue(), "expected first emitter to be initialized")
			Expect(mock2.initCalled).To(BeTrue(), "expected second emitter to be initialized")
		})

		It("handles initialization error", func() {
			mock2.initErr = errors.New("init failed")
			Expect(emitter.Init()).To(HaveOccurred())
		})

		It("works with no emitters", func() {
			emitter = fanout.NewEmitter()
			Expect(emitter.Init()).To(Succeed())
		})
	})

	Context("when emitting entries", func() {
		It("broadcasts entries to all emitters", func() {
			testEntry := &entry.Entry{
				Unstructured: "test log",
			}
			emitter.Emit(testEntry)

			Expect(mock1.entries).To(HaveLen(1))
			Expect(mock2.entries).To(HaveLen(1))
			Expect(mock1.entries[0]).To(Equal(testEntry))
			Expect(mock2.entries[0]).To(Equal(testEntry))
		})

		It("does not panic with no emitters", func() {
			emitter = fanout.NewEmitter()
			Expect(func() {
				emitter.Emit(&entry.Entry{Unstructured: "test"})
			}).NotTo(Panic())
		})
	})
})
