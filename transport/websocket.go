package transport

import (
	"context"
	"log/slog"
)

// WebSocketTransport connects to the Flagmint backend over a persistent
// WebSocket connection and streams flag updates in real time.
//
// NOTE: Full implementation is tracked in Ticket 3.
type WebSocketTransport struct {
	endpoint  string
	apiKey    string
	logger    *slog.Logger
	onUpdated func(flags map[string]any)
}

// NewWebSocketTransport creates a new WebSocketTransport.
func NewWebSocketTransport(endpoint, apiKey string, logger *slog.Logger) *WebSocketTransport {
	return &WebSocketTransport{
		endpoint: endpoint,
		apiKey:   apiKey,
		logger:   logger,
	}
}

// Connect establishes the WebSocket connection.
func (t *WebSocketTransport) Connect(_ context.Context) error {
	t.logger.Info("websocket transport: connect (stub)", "endpoint", t.endpoint)
	return nil
}

// Close shuts down the WebSocket connection.
func (t *WebSocketTransport) Close() error {
	return nil
}

// OnFlagsUpdated registers the update callback.
func (t *WebSocketTransport) OnFlagsUpdated(fn func(flags map[string]any)) {
	t.onUpdated = fn
}
