package websocket_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	wsemitter "github.com/appthrust/kutelog/pkg/emitters/websocket"
	"github.com/appthrust/kutelog/pkg/entry"
)

var _ = Describe("WebSocket Emitter", func() {
	var (
		emitter *wsemitter.Emitter
		wsURL   string
	)

	BeforeEach(func() {
		emitter = wsemitter.NewEmitter()
		Expect(emitter.Init()).To(Succeed())

		// Get WebSocket URL
		addr := emitter.Address()
		wsURL = "ws://" + strings.TrimPrefix(addr, "http://") + "/ws"
	})

	Context("when initializing", func() {
		It("starts server on a free port", func() {
			Expect(emitter.Address()).NotTo(BeEmpty())
		})

		It("serves viewer page", func() {
			resp, err := http.Get(emitter.Address())
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(resp.Header.Get("Content-Type")).To(Equal("text/html"))
		})
	})

	Context("when handling WebSocket connections", func() {
		It("accepts client connection", func() {
			ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			Expect(err).NotTo(HaveOccurred())
			defer ws.Close()
		})

		It("broadcasts structured log", func() {
			ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			Expect(err).NotTo(HaveOccurred())
			defer ws.Close()

			timestamp := time.Now()
			emitter.Emit(&entry.Entry{
				Structured: &entry.Structured{
					Timestamp: timestamp,
					Level:     entry.LevelInfo,
					Message:   "test message",
					Data: map[string]interface{}{
						"key": "value",
					},
				},
			})

			_, message, err := ws.ReadMessage()
			Expect(err).NotTo(HaveOccurred())

			var msg wsemitter.Message
			Expect(json.Unmarshal(message, &msg)).To(Succeed())
			received, ok := msg.Body.(map[string]interface{})
			Expect(ok).To(BeTrue())

			receivedTime, err := time.Parse(time.RFC3339Nano, received["timestamp"].(string))
			Expect(err).NotTo(HaveOccurred())
			Expect(receivedTime).To(BeTemporally("~", timestamp, time.Second))
			Expect(received["level"]).To(Equal("info"))
			Expect(received["message"]).To(Equal("test message"))
			Expect(received["data"].(map[string]interface{})["key"]).To(Equal("value"))
		})

		It("broadcasts unstructured log", func() {
			ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			Expect(err).NotTo(HaveOccurred())
			defer ws.Close()

			emitter.Emit(&entry.Entry{
				Unstructured: "plain text log",
			})

			_, message, err := ws.ReadMessage()
			Expect(err).NotTo(HaveOccurred())
			var msg wsemitter.Message
			Expect(json.Unmarshal(message, &msg)).To(Succeed())
			received, ok := msg.Body.(string)
			Expect(ok).To(BeTrue())
			Expect(received).To(Equal("plain text log"))
		})

		It("handles multiple clients", func() {
			ws1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			Expect(err).NotTo(HaveOccurred())
			defer ws1.Close()

			ws2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			Expect(err).NotTo(HaveOccurred())
			defer ws2.Close()

			emitter.Emit(&entry.Entry{
				Unstructured: "broadcast test",
			})

			_, message1, err := ws1.ReadMessage()
			Expect(err).NotTo(HaveOccurred())
			var msg1 wsemitter.Message
			Expect(json.Unmarshal(message1, &msg1)).To(Succeed())
			received1, ok := msg1.Body.(string)
			Expect(ok).To(BeTrue())
			Expect(received1).To(Equal("broadcast test"))

			_, message2, err := ws2.ReadMessage()
			Expect(err).NotTo(HaveOccurred())
			var msg2 wsemitter.Message
			Expect(json.Unmarshal(message2, &msg2)).To(Succeed())
			received2, ok := msg2.Body.(string)
			Expect(ok).To(BeTrue())
			Expect(received2).To(Equal("broadcast test"))
		})

		It("removes client on disconnection", func() {
			ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			Expect(err).NotTo(HaveOccurred())

			ws.Close()

			// Should not panic when broadcasting after client disconnection
			emitter.Emit(&entry.Entry{
				Unstructured: "after disconnect",
			})
		})

		It("sends message history to new clients", func() {
			// Send two messages before connecting the second client
			emitter.Emit(&entry.Entry{
				Unstructured: "message 1",
			})
			emitter.Emit(&entry.Entry{
				Unstructured: "message 2",
			})

			// Connect new client
			ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			Expect(err).NotTo(HaveOccurred())
			defer ws.Close()

			// Should receive both messages from history
			for i, expected := range []string{"message 1", "message 2"} {
				_, message, err := ws.ReadMessage()
				Expect(err).NotTo(HaveOccurred())
				var msg wsemitter.Message
				Expect(json.Unmarshal(message, &msg)).To(Succeed())
				received, ok := msg.Body.(string)
				Expect(ok).To(BeTrue(), "Message %d should be a string", i+1)
				Expect(received).To(Equal(expected))
			}
		})
	})

	Context("when serving HTTP endpoints", func() {
		It("serves version information", func() {
			resp, err := http.Get(emitter.Address() + "/version")
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))

			var data map[string]string
			Expect(json.NewDecoder(resp.Body).Decode(&data)).To(Succeed())
			Expect(data).To(HaveKey("name"))
			Expect(data).To(HaveKey("version"))
		})

		It("returns 404 for non-existent files", func() {
			resp, err := http.Get(emitter.Address() + "/nonexistent")
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("serves index page with correct Content-Type", func() {
			resp, err := http.Get(emitter.Address() + "/")
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(resp.Header.Get("Content-Type")).To(Equal("text/html"))
		})
	})
})
