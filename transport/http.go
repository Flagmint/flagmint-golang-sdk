package transport

import (
	"context"
	"log/slog"
)

// HTTPTransport polls the Flagmint backend periodically over HTTPS (long-polling).
//
// NOTE: Full implementation is tracked in Ticket 3.
type HTTPTransport struct {
	endpoint  string
	apiKey    string
	logger    *slog.Logger
	onUpdated func(flags map[string]any)
}

// NewHTTPTransport creates a new HTTPTransport.
func NewHTTPTransport(endpoint, apiKey string, logger *slog.Logger) *HTTPTransport {
	return &HTTPTransport{
		endpoint: endpoint,
		apiKey:   apiKey,
		logger:   logger,
	}
}

// Connect starts the long-polling loop.
func (t *HTTPTransport) Connect(_ context.Context) error {
	t.logger.Info("http transport: connect (stub)", "endpoint", t.endpoint)
	return nil
}

// Close stops the long-polling loop.
func (t *HTTPTransport) Close() error {
	return nil
}

// OnFlagsUpdated registers the update callback.
func (t *HTTPTransport) OnFlagsUpdated(fn func(flags map[string]any)) {
	t.onUpdated = fn
}
