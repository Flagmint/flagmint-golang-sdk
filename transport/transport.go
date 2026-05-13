// Package transport defines the Transport interface and its implementations.
package transport

import "context"

// Transport is responsible for receiving flag updates from the Flagmint
// backend and delivering them to the client.
type Transport interface {
	// Connect establishes the connection to the backend.
	// It blocks until the connection is established or ctx is cancelled.
	Connect(ctx context.Context) error

	// Close shuts down the transport and releases all associated resources.
	Close() error

	// OnFlagsUpdated registers a callback that is invoked whenever a new set
	// of flags is received from the backend. The callback must not block.
	OnFlagsUpdated(fn func(flags map[string]any))
}
