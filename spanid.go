package nop

import (
	"github.com/bassosimone/runtimex"
	"github.com/google/uuid"
)

// NewSpanID returns a UUIDv7 representing a span.
//
// A span is a sequence of operation that can fail in a single, specific
// way. For example, a workflow to perform a TLS handshake with an endpoint
// or a single DNS-over-HTTPS exchange with an endpoint.
//
// We recommend using a span ID for uniquely identifying spans.
//
// The span terminology is borrowed from OTel.
//
// This function panics if the system random number generator fails,
// which should only happen under extraordinary circumstances.
func NewSpanID() string {
	return runtimex.PanicOnError1(uuid.NewV7()).String()
}
