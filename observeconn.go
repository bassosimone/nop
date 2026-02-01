//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Adapted from: https://github.com/ooni/probe-cli/blob/v3.20.1/internal/measurexlite/conn.go
// Adapted from: https://github.com/rbmk-project/rbmk/blob/v0.17.0/pkg/x/netcore/conn.go
//

package nop

import (
	"context"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/bassosimone/safeconn"
)

// NewObserveConnFunc returns a new [*ObserveConnFunc] with default logging.
//
// The cfg argument contains the common configuration for nop operations.
//
// The logger argument is the [SLogger] to use for structured logging.
func NewObserveConnFunc(cfg *Config, logger SLogger) *ObserveConnFunc {
	return &ObserveConnFunc{
		ErrClassifier: cfg.ErrClassifier,
		Logger:        logger,
		TimeNow:       cfg.TimeNow,
	}
}

// ObserveConnFunc observes a [net.Conn] to log I/O operations.
//
// This primitive provides observability for network operations by logging
// all I/O events including reads, writes, and deadline changes. For timeout
// enforcement, use [CancelWatchFunc] to close the connection when the context
// is done, which causes any in-progress I/O to fail immediately.
//
// All fields are safe to modify after construction but before first use.
// Fields must not be mutated concurrently with calls to [Call].
type ObserveConnFunc struct {
	// ErrClassifier classifies errors for structured logging.
	//
	// Set by [NewObserveConnFunc] from [Config.ErrClassifier].
	ErrClassifier ErrClassifier

	// Logger is the [SLogger] to use (configurable for testing or custom logging).
	//
	// Set by [NewObserveConnFunc] to the user-provided logger.
	Logger SLogger

	// TimeNow is the function to get the current time (configurable for testing).
	//
	// Set by [NewObserveConnFunc] from [Config.TimeNow].
	TimeNow func() time.Time
}

var _ Func[net.Conn, net.Conn] = &ObserveConnFunc{}

// Call invokes the [*ObserveConnFunc] to observe a [net.Conn] for logging I/O operations.
func (op *ObserveConnFunc) Call(ctx context.Context, conn net.Conn) (net.Conn, error) {
	observed := &observedConn{
		closeonce: sync.Once{},
		conn:      conn,
		laddr:     safeconn.LocalAddr(conn),
		op:        op,
		protocol:  safeconn.Network(conn),
		raddr:     safeconn.RemoteAddr(conn),
	}
	return observed, nil
}

// observedConn observes a [net.Conn].
type observedConn struct {
	closeonce sync.Once
	conn      net.Conn
	laddr     string
	op        *ObserveConnFunc
	protocol  string
	raddr     string
}

// Close implements [net.Conn].
//
// Subsequent calls return [net.ErrClosed], consistent with Go's standard
// library behavior for closed connections.
func (c *observedConn) Close() (err error) {
	err = net.ErrClosed
	c.closeonce.Do(func() {
		t0 := c.op.TimeNow()
		c.op.Logger.Info(
			"closeStart",
			slog.String("localAddr", c.laddr),
			slog.String("protocol", c.protocol),
			slog.String("remoteAddr", c.raddr),
			slog.Time("t", t0),
		)

		err = c.conn.Close()

		c.op.Logger.Info(
			"closeDone",
			slog.Any("err", err),
			slog.String("errClass", c.op.ErrClassifier.Classify(err)),
			slog.String("localAddr", c.laddr),
			slog.String("protocol", c.protocol),
			slog.String("remoteAddr", c.raddr),
			slog.Time("t0", t0),
			slog.Time("t", c.op.TimeNow()),
		)
	})
	return
}

// LocalAddr implements [net.Conn].
func (c *observedConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// Read implements [net.Conn].
func (c *observedConn) Read(buf []byte) (int, error) {
	t0 := c.op.TimeNow()
	c.op.Logger.Debug(
		"readStart",
		slog.Int("ioBufferSize", len(buf)),
		slog.String("localAddr", c.laddr),
		slog.String("protocol", c.protocol),
		slog.String("remoteAddr", c.raddr),
		slog.Time("t", t0),
	)

	count, err := c.conn.Read(buf)

	c.op.Logger.Debug(
		"readDone",
		slog.Int("ioBytesCount", count),
		slog.Any("err", err),
		slog.String("errClass", c.op.ErrClassifier.Classify(err)),
		slog.String("localAddr", c.laddr),
		slog.String("protocol", c.protocol),
		slog.String("remoteAddr", c.raddr),
		slog.Time("t0", t0),
		slog.Time("t", c.op.TimeNow()),
	)

	return count, err
}

// RemoteAddr implements [net.Conn].
func (c *observedConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// SetDeadline implements [net.Conn].
func (c *observedConn) SetDeadline(t time.Time) error {
	c.op.Logger.Debug(
		"setDeadline",
		slog.Time("deadline", t),
		slog.String("localAddr", c.laddr),
		slog.String("protocol", c.protocol),
		slog.String("remoteAddr", c.raddr),
		slog.Time("t", c.op.TimeNow()),
	)
	return c.conn.SetDeadline(t)
}

// SetReadDeadline implements [net.Conn].
func (c *observedConn) SetReadDeadline(t time.Time) error {
	c.op.Logger.Debug(
		"setReadDeadline",
		slog.Time("deadline", t),
		slog.String("localAddr", c.laddr),
		slog.String("protocol", c.protocol),
		slog.String("remoteAddr", c.raddr),
		slog.Time("t", c.op.TimeNow()),
	)
	return c.conn.SetReadDeadline(t)
}

// SetWriteDeadline implements [net.Conn].
func (c *observedConn) SetWriteDeadline(t time.Time) error {
	c.op.Logger.Debug(
		"setWriteDeadline",
		slog.Time("deadline", t),
		slog.String("localAddr", c.laddr),
		slog.String("protocol", c.protocol),
		slog.String("remoteAddr", c.raddr),
		slog.Time("t", c.op.TimeNow()),
	)
	return c.conn.SetWriteDeadline(t)
}

// Write implements [net.Conn].
func (c *observedConn) Write(data []byte) (n int, err error) {
	t0 := c.op.TimeNow()
	c.op.Logger.Debug(
		"writeStart",
		slog.Int("ioBufferSize", len(data)),
		slog.String("localAddr", c.laddr),
		slog.String("protocol", c.protocol),
		slog.String("remoteAddr", c.raddr),
		slog.Time("t", t0),
	)

	count, err := c.conn.Write(data)

	c.op.Logger.Debug(
		"writeDone",
		slog.Int("ioBytesCount", count),
		slog.Any("err", err),
		slog.String("errClass", c.op.ErrClassifier.Classify(err)),
		slog.String("localAddr", c.laddr),
		slog.String("protocol", c.protocol),
		slog.String("remoteAddr", c.raddr),
		slog.Time("t0", t0),
		slog.Time("t", c.op.TimeNow()),
	)

	return count, err
}
