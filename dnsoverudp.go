// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"context"
	"net"
	"net/netip"
	"time"

	"github.com/bassosimone/dnscodec"
	"github.com/bassosimone/minest"
	"github.com/bassosimone/safeconn"
)

// DNSOverUDPConn wraps a UDP connection for DNS-over-UDP exchanges.
//
// This type owns the underlying connection. The caller is responsible for
// calling Close() when done.
//
// All fields are safe to modify after construction but before first use of
// Exchange(). Fields must not be mutated concurrently with Exchange().
//
// Construct via [*DNSOverUDPConnFunc].
type DNSOverUDPConn struct {
	// conn is the owned UDP connection.
	conn net.Conn

	// ErrClassifier classifies errors for structured logging.
	ErrClassifier ErrClassifier

	// Logger is the SLogger to use.
	Logger SLogger

	// TimeNow is the function to get the current time.
	TimeNow func() time.Time
}

// Close closes the underlying UDP connection.
func (c *DNSOverUDPConn) Close() error {
	return c.conn.Close()
}

// Conn returns the underlying net.Conn for logging purposes.
func (c *DNSOverUDPConn) Conn() net.Conn {
	return c.conn
}

// Exchange performs a DNS exchange over UDP.
// This method may be called multiple times on the same connection.
func (c *DNSOverUDPConn) Exchange(ctx context.Context, query *dnscodec.Query) (*dnscodec.Response, error) {
	// 1. Get the owned connection
	conn := c.conn

	// 2. Create the log context
	t0 := c.TimeNow()
	deadline, _ := ctx.Deadline()
	var rqr []byte
	lc := &dnsExchangeLogContext{
		ErrClassifier:  c.ErrClassifier,
		LocalAddr:      safeconn.LocalAddr(conn),
		Logger:         c.Logger,
		Protocol:       safeconn.Network(conn),
		RemoteAddr:     safeconn.RemoteAddr(conn),
		ServerProtocol: "udp",
		TimeNow:        c.TimeNow,
	}

	// 3. Create the transport
	//
	// Note: we're not going to dial, so let's use a dialer that panics
	// if we attempt to dial (programmer error).
	txp := minest.NewDNSOverUDPTransport(dnsUnusedDialer{}, netip.AddrPortFrom(netip.IPv4Unspecified(), 0))

	// 4. Set observers for raw messages
	txp.ObserveRawQuery = lc.makeQueryObserver(t0, &rqr)
	txp.ObserveRawResponse = lc.makeResponseObserver(t0, &rqr)

	// 5. Execute with logging
	lc.logStart(t0, deadline)
	resp, err := txp.ExchangeWithConn(ctx, conn, query)
	lc.logDone(t0, deadline, err)

	return resp, err
}

// DNSOverUDPConnFunc wraps a net.Conn into a [*DNSOverUDPConn].
//
// This is a [Func] that can be composed into pipelines.
//
// All fields are safe to modify after construction but before first use.
// Fields must not be mutated concurrently with calls to [Call].
type DNSOverUDPConnFunc struct {
	// ErrClassifier classifies errors for structured logging.
	//
	// Set by [NewDNSOverUDPConnFunc] from [Config.ErrClassifier].
	ErrClassifier ErrClassifier

	// Logger is the [SLogger] to use (configurable for testing or custom logging).
	//
	// Set by [NewDNSOverUDPConnFunc] to the user-provided logger.
	Logger SLogger

	// TimeNow is the function to get the current time (configurable for testing).
	//
	// Set by [NewDNSOverUDPConnFunc] from [Config.TimeNow].
	TimeNow func() time.Time
}

// NewDNSOverUDPConnFunc returns a new [*DNSOverUDPConnFunc].
//
// The cfg argument contains the common configuration for nop operations.
//
// The logger argument is the [SLogger] to use for structured logging.
func NewDNSOverUDPConnFunc(cfg *Config, logger SLogger) *DNSOverUDPConnFunc {
	return &DNSOverUDPConnFunc{
		ErrClassifier: cfg.ErrClassifier,
		Logger:        logger,
		TimeNow:       cfg.TimeNow,
	}
}

var _ Func[net.Conn, *DNSOverUDPConn] = &DNSOverUDPConnFunc{}

// Call wraps the net.Conn into a DNSOverUDPConn.
func (op *DNSOverUDPConnFunc) Call(ctx context.Context, conn net.Conn) (*DNSOverUDPConn, error) {
	return &DNSOverUDPConn{
		conn:          conn,
		ErrClassifier: op.ErrClassifier,
		Logger:        op.Logger,
		TimeNow:       op.TimeNow,
	}, nil
}
