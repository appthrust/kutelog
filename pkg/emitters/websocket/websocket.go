package websocket

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/appthrust/kutelog/pkg/core"
	"github.com/appthrust/kutelog/pkg/entry"
	"github.com/appthrust/kutelog/pkg/version"
	"github.com/gorilla/websocket"
)

//go:embed static/dist/*
var staticFiles embed.FS

// DefaultPort is the default port for the WebSocket server
const DefaultPort = 9106

var _ core.Emitter = &Emitter{}

// Emitter implements WebSocket server that broadcasts log entries to connected clients
// Message represents a WebSocket message with ID
type Message struct {
	ID   int64       `json:"id"` // Combination of timestamp and sequence number (see Emitter.sequence for details)
	Body interface{} `json:"body"`
}

type Emitter struct {
	server         *http.Server
	upgrader       websocket.Upgrader
	clients        sync.Map     // map[*websocket.Conn]struct{}
	addr           string       // server address for testing
	messageHistory []Message    // stores message history for replay
	historyMutex   sync.RWMutex // synchronizes access to message history

	// Sequence number (12 bits, 0-4095)
	// Due to JavaScript Number type's 53-bit precision limitation, we use the following bit allocation:
	// - 41 bits: Unix timestamp in milliseconds (supports dates until year 2286)
	// - 12 bits: Sequence number (allows unique identification of up to 4,096 messages per millisecond)
	// Total: 53 bits, ensuring the value stays within Number.MAX_SAFE_INTEGER (2^53-1) in JavaScript
	sequence atomic.Int64
}

// NewEmitter creates a new WebSocket emitter
func NewEmitter() *Emitter {
	return &Emitter{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // allow all origins for development
			},
		},
		clients:        sync.Map{},         // sync.Map is a zero value, no need to initialize
		messageHistory: make([]Message, 0), // initialize message history
	}
}

// isPortInUseError checks if the error is due to the port being in use
func isPortInUseError(err error) bool {
	if opErr, ok := err.(*net.OpError); ok {
		if syscallErr, ok := opErr.Err.(*os.SyscallError); ok {
			return syscallErr.Err == syscall.EADDRINUSE
		}
	}
	return false
}

// Init starts WebSocket server on port 9106 or the next available port
func (e *Emitter) Init() error {
	// Try to listen on default port first
	port := DefaultPort
	var listener net.Listener
	var err error

	// Try ports until we find an available one
	for {
		listener, err = net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			break
		}
		if !isPortInUseError(err) {
			return fmt.Errorf("failed to listen: %w", err)
		}
		port++
	}

	e.addr = listener.Addr().String()

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", e.handleWS)
	mux.HandleFunc("/", e.handleIndex)
	mux.HandleFunc("/version", e.handleVersion)

	// Start server
	e.server = &http.Server{
		Handler: mux,
	}
	fmt.Printf("WebSocket server listening on http://%s\n", e.addr)
	go e.server.Serve(listener)
	return nil
}

// Address returns server address (for testing)
func (e *Emitter) Address() string {
	return "http://" + e.addr
}

// handleWS handles WebSocket connections
func (e *Emitter) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := e.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// Send message history to new client
	e.historyMutex.RLock()
	for _, msg := range e.messageHistory {
		data, err := json.Marshal(msg)
		if err != nil {
			e.historyMutex.RUnlock()
			conn.Close()
			return
		}
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			e.historyMutex.RUnlock()
			conn.Close()
			return
		}
	}
	e.historyMutex.RUnlock()

	e.clients.Store(conn, struct{}{})
	go func() {
		defer func() {
			conn.Close()
			e.clients.Delete(conn)
		}()
		// Wait for client disconnection
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
}

// handleVersion serves version information
func (e *Emitter) handleVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"name":    version.Name,
		"version": version.Version,
	})
}

// handleIndex serves static files
func (e *Emitter) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	content, err := fs.ReadFile(staticFiles, "static/dist/index.html")
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(content)
}

// Emit sends log entry to all connected clients
func (e *Emitter) Emit(entry *entry.Entry) {
	if entry == nil {
		return
	}

	var data []byte
	var err error
	if entry.Structured != nil {
		// Encode structured log as JSON
		if data, err = json.Marshal(entry.Structured); err != nil {
			return
		}
	} else if entry.Unstructured != "" {
		// Send unstructured log as JSON string
		if data, err = json.Marshal(entry.Unstructured); err != nil {
			return
		}
	} else {
		return
	}

	// Create message with timestamp and sequence number
	e.historyMutex.Lock()
	currentTime := time.Now().UnixMilli()
	seq := e.sequence.Add(1) & 0xFFF // Use lower 12 bits (cycles through 0-4095)
	msg := Message{
		// Message ID generation:
		// 1. Shift timestamp left by 12 bits to use the upper 41 bits
		// 2. Use sequence number in the lower 12 bits (cycles through 0-4095)
		// This ensures the ID stays within 53 bits for safe handling in JavaScript clients
		// and provides unique IDs for up to 4,096 messages within the same millisecond
		ID:   (currentTime << 12) | seq,
		Body: entry.Structured,
	}
	if entry.Structured == nil {
		msg.Body = entry.Unstructured
	}

	// Store message in history
	e.messageHistory = append(e.messageHistory, msg)
	e.historyMutex.Unlock()

	// Marshal message to JSON
	data, err = json.Marshal(msg)
	if err != nil {
		return
	}

	// Broadcast to all clients
	e.clients.Range(func(key, _ interface{}) bool {
		conn := key.(*websocket.Conn)
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			conn.Close()
			e.clients.Delete(conn)
		}
		return true
	})
}

// Close shuts down the WebSocket server
func (e *Emitter) Close() error {
	if e.server != nil {
		return e.server.Close()
	}
	return nil
}
