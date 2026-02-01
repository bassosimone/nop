// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"context"
	"net"
)

// NewCancelWatchFunc returns a new [*CancelWatchFunc].
func NewCancelWatchFunc() *CancelWatchFunc {
	return &CancelWatchFunc{}
}

// CancelWatchFunc arranges for the connection to be closed when the context
// is done (cancelled or deadline exceeded). This provides responsive cleanup
// on external cancellation (e.g., SIGINT via signal.NotifyContext) rather than
// waiting for per-operation timeouts.
//
// The returned connection wraps the input connection. Closing the returned
// connection unregisters the context watcher and closes the underlying
// connection. This ensures no goroutine leaks even if the context is
// never cancelled.
//
// The watcher is safe to use with any [net.Conn] implementation because
// Go's standard library uses the [net.ErrClosed] pattern: closing an
// already-closed connection returns [net.ErrClosed], and I/O operations
// on a closed connection fail gracefully. The [ObserveConnFunc] wrapper
// follows this same pattern.
//
// Use this primitive in pipelines where:
//   - The context lifetime matches the intended connection lifetime
//   - Immediate cleanup on cancellation is desired (e.g., CLI tools)
//
// Do not use this primitive when:
//   - The connection will be returned and may outlive the current context
//   - You're implementing a connection pool or long-lived connection management
type CancelWatchFunc struct{}

var _ Func[net.Conn, net.Conn] = &CancelWatchFunc{}

// Call registers a context watcher using [context.AfterFunc] that closes
// the connection when the context is done. The returned [net.Conn] wraps
// the input: closing it unregisters the watcher and closes the underlying
// connection.
func (op *CancelWatchFunc) Call(ctx context.Context, conn net.Conn) (net.Conn, error) {
	stop := context.AfterFunc(ctx, func() {
		conn.Close()
	})
	return &cancelWatchedConn{Conn: conn, stop: stop}, nil
}

// cancelWatchedConn wraps a [net.Conn] with a context cancellation watcher.
type cancelWatchedConn struct {
	net.Conn
	stop func() bool
}

// Close unregisters the context watcher and closes the underlying connection.
func (c *cancelWatchedConn) Close() error {
	c.stop()
	return c.Conn.Close()
}
