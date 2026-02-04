// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"context"
	"net"
	"net/netip"
	"time"

	"github.com/bassosimone/dnscodec"
	"github.com/bassosimone/dnsoverstream"
	"github.com/bassosimone/safeconn"
)

// DNSOverTCPConn wraps a TCP connection for DNS-over-TCP exchanges.
//
// This type owns the underlying connection. The caller is responsible for
// calling Close() when done.
//
// All fields are safe to modify after construction but before first use of
// Exchange(). Fields must not be mutated concurrently with Exchange().
//
// Construct via [*DNSOverTCPConnFunc].
type DNSOverTCPConn struct {
	// conn is the owned TCP connection.
	conn net.Conn

	// ErrClassifier classifies errors for structured logging.
	ErrClassifier ErrClassifier

	// Logger is the SLogger to use.
	Logger SLogger

	// TimeNow is the function to get the current time.
	TimeNow func() time.Time
}

// Close closes the underlying TCP connection.
func (c *DNSOverTCPConn) Close() error {
	return c.conn.Close()
}

// Conn returns the underlying net.Conn for logging purposes.
func (c *DNSOverTCPConn) Conn() net.Conn {
	return c.conn
}

// Exchange performs a DNS exchange over TCP.
// This method may be called multiple times on the same connection.
func (c *DNSOverTCPConn) Exchange(ctx context.Context, query *dnscodec.Query) (*dnscodec.Response, error) {
	// 1. Get the owned connection
	conn := c.conn

	// 2. Create the log context
	t0 := c.TimeNow()
	deadline, _ := ctx.Deadline()
	var rqr []byte
	lc := &DNSExchangeLogContext{
		ErrClassifier:  c.ErrClassifier,
		LocalAddr:      safeconn.LocalAddr(conn),
		Logger:         c.Logger,
		Protocol:       safeconn.Network(conn),
		RemoteAddr:     safeconn.RemoteAddr(conn),
		ServerProtocol: "tcp",
		TimeNow:        c.TimeNow,
	}

	// 3. Create the transport
	//
	// Note: we're not going to dial, so let's use a dialer that panics
	// if we attempt to dial (programmer error).
	streamDialer := dnsoverstream.NewStreamOpenerDialerTCP(dnsUnusedDialer{})
	txp := dnsoverstream.NewTransport(streamDialer, netip.AddrPortFrom(netip.IPv4Unspecified(), 0))

	// 4. Set observers for raw messages
	txp.ObserveRawQuery = lc.MakeQueryObserver(t0, &rqr)
	txp.ObserveRawResponse = lc.MakeResponseObserver(t0, &rqr)

	// 5. Execute with logging
	lc.LogStart(t0, deadline)
	so := dnsoverstream.NewTCPStreamOpener(conn)
	resp, err := txp.ExchangeWithStreamOpener(ctx, so, query)
	lc.LogDone(t0, deadline, err)

	return resp, err
}

// DNSOverTCPConnFunc wraps a net.Conn into a [*DNSOverTCPConn].
//
// This is a [Func] that can be composed into pipelines.
//
// All fields are safe to modify after construction but before first use.
// Fields must not be mutated concurrently with calls to [Call].
type DNSOverTCPConnFunc struct {
	// ErrClassifier classifies errors for structured logging.
	//
	// Set by [NewDNSOverTCPConnFunc] from [Config.ErrClassifier].
	ErrClassifier ErrClassifier

	// Logger is the [SLogger] to use (configurable for testing or custom logging).
	//
	// Set by [NewDNSOverTCPConnFunc] to the user-provided logger.
	Logger SLogger

	// TimeNow is the function to get the current time (configurable for testing).
	//
	// Set by [NewDNSOverTCPConnFunc] from [Config.TimeNow].
	TimeNow func() time.Time
}

// NewDNSOverTCPConnFunc returns a new [*DNSOverTCPConnFunc].
//
// The cfg argument contains the common configuration for nop operations.
//
// The logger argument is the [SLogger] to use for structured logging.
func NewDNSOverTCPConnFunc(cfg *Config, logger SLogger) *DNSOverTCPConnFunc {
	return &DNSOverTCPConnFunc{
		ErrClassifier: cfg.ErrClassifier,
		Logger:        logger,
		TimeNow:       cfg.TimeNow,
	}
}

var _ Func[net.Conn, *DNSOverTCPConn] = &DNSOverTCPConnFunc{}

// Call wraps the net.Conn into a DNSOverTCPConn.
func (op *DNSOverTCPConnFunc) Call(ctx context.Context, conn net.Conn) (*DNSOverTCPConn, error) {
	return &DNSOverTCPConn{
		conn:          conn,
		ErrClassifier: op.ErrClassifier,
		Logger:        op.Logger,
		TimeNow:       op.TimeNow,
	}, nil
}
