package flagmint

import (
	"log/slog"

	"github.com/flagmint/flagmint-go/cache"
)

// TransportMode controls the transport mechanism used by the client.
type TransportMode string

const (
	// TransportAuto selects WebSocket when available, falling back to long-polling.
	TransportAuto TransportMode = "auto"
	// TransportWebSocket forces the WebSocket transport.
	TransportWebSocket TransportMode = "websocket"
	// TransportLongPolling forces the HTTP long-polling transport.
	TransportLongPolling TransportMode = "long-polling"
)

// Option configures the FlagClient. Use With* functions to create options.
type Option func(*clientConfig)

type clientConfig struct {
	apiKey        string
	context       *EvaluationContext
	transportMode TransportMode
	enableCache   bool
	cacheAdapter  cache.Adapter
	onError       func(error)
	restEndpoint  string
	wsEndpoint    string
	deferInit     bool
	logger        *slog.Logger
}

func defaultConfig() clientConfig {
	return clientConfig{
		transportMode: TransportAuto,
		restEndpoint:  "https://api.flagmint.com",
		wsEndpoint:    "wss://api.flagmint.com",
		logger:        slog.Default(),
	}
}

// WithContext sets the default evaluation context for the client.
func WithContext(ctx EvaluationContext) Option {
	return func(c *clientConfig) {
		c.context = &ctx
	}
}

// WithTransportMode sets the transport mechanism.
func WithTransportMode(mode TransportMode) Option {
	return func(c *clientConfig) {
		c.transportMode = mode
	}
}

// WithCache enables or disables the flag cache.
func WithCache(enabled bool) Option {
	return func(c *clientConfig) {
		c.enableCache = enabled
	}
}

// WithCacheAdapter sets a custom cache adapter.
func WithCacheAdapter(adapter cache.Adapter) Option {
	return func(c *clientConfig) {
		c.cacheAdapter = adapter
		c.enableCache = true
	}
}

// WithOnError registers a callback invoked when the client encounters a
// non-fatal error (e.g., a failed flag refresh).
func WithOnError(fn func(error)) Option {
	return func(c *clientConfig) {
		c.onError = fn
	}
}

// WithEndpoints overrides the default REST and WebSocket API endpoints.
func WithEndpoints(rest, ws string) Option {
	return func(c *clientConfig) {
		c.restEndpoint = rest
		c.wsEndpoint = ws
	}
}

// WithDeferInit prevents the client from connecting immediately on creation.
// Call [FlagClient.Initialize] manually when ready.
func WithDeferInit() Option {
	return func(c *clientConfig) {
		c.deferInit = true
	}
}

// WithLogger sets the structured logger used by the client.
func WithLogger(l *slog.Logger) Option {
	return func(c *clientConfig) {
		c.logger = l
	}
}
