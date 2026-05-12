// Package flagmint is the Go SDK for the Flagmint feature flag service.
//
// # Quick start
//
//	client, err := flagmint.NewClient("your-api-key",
//	    flagmint.WithContext(flagmint.EvaluationContext{
//	        Kind: "user",
//	        Key:  "user-123",
//	    }),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	if val, ok := client.GetFlag("my-feature"); ok && val.(bool) {
//	    // feature is enabled
//	}
//
// # Architecture
//
// The SDK maintains a persistent connection to the Flagmint backend (WebSocket
// by default, with HTTP long-polling as a fallback).  Flags are evaluated
// server-side and streamed to the client in real time.  An optional local
// evaluator (Ticket 5) can evaluate flags without a round-trip when offline.
package flagmint
