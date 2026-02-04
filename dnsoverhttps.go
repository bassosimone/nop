// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"context"
	"time"

	"github.com/bassosimone/dnscodec"
	"github.com/bassosimone/dnsoverhttps"
	"github.com/bassosimone/safeconn"
)

// DNSOverHTTPSConn wraps an HTTPConn for DNS-over-HTTPS exchanges.
//
// This type owns the underlying HTTPConn. The caller is responsible for
// calling Close() when done.
//
// All fields are safe to modify after construction but before first use of
// Exchange(). Fields must not be mutated concurrently with Exchange().
//
// Construct via [*DNSOverHTTPSConnFunc].
type DNSOverHTTPSConn struct {
	// httpConn is the owned HTTPConn.
	httpConn *HTTPConn

	// url is the DoH endpoint URL.
	url string

	// ErrClassifier classifies errors for structured logging.
	ErrClassifier ErrClassifier

	// Logger is the SLogger to use.
	Logger SLogger

	// TimeNow is the function to get the current time.
	TimeNow func() time.Time
}

// Close closes the underlying HTTPConn.
func (c *DNSOverHTTPSConn) Close() error {
	return c.httpConn.Close()
}

// HTTPConn returns the underlying *HTTPConn for logging purposes.
func (c *DNSOverHTTPSConn) HTTPConn() *HTTPConn {
	return c.httpConn
}

// Exchange performs a DNS exchange over HTTPS.
// This method may be called multiple times on the same connection.
func (c *DNSOverHTTPSConn) Exchange(ctx context.Context, query *dnscodec.Query) (*dnscodec.Response, error) {
	// 1. Get the owned HTTPConn and underlying connection for logging
	hc := c.httpConn
	conn := hc.Conn()

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
		ServerProtocol: "doh",
		TimeNow:        c.TimeNow,
	}

	// 3. Create the HTTP request and the query message
	lc.LogStart(t0, deadline)
	httpReq, queryMsg, err := dnsoverhttps.NewRequestWithHook(ctx, query, c.url, lc.MakeQueryObserver(t0, &rqr))
	if err != nil {
		lc.LogDone(t0, deadline, err)
		return nil, err
	}

	// 4. Perform the HTTP round trip
	httpResp, err := hc.RoundTrip(httpReq)
	if err != nil {
		lc.LogDone(t0, deadline, err)
		return nil, err
	}

	// 5. Read the response and validate it
	resp, err := dnsoverhttps.ReadResponseWithHook(ctx, httpResp, queryMsg, lc.MakeResponseObserver(t0, &rqr))
	lc.LogDone(t0, deadline, err)
	return resp, err
}

// DNSOverHTTPSConnFunc wraps an *HTTPConn into a [*DNSOverHTTPSConn].
//
// This is a [Func] that can be composed into pipelines.
//
// All fields are safe to modify after construction but before first use.
// Fields must not be mutated concurrently with calls to [Call].
type DNSOverHTTPSConnFunc struct {
	// URL is the DoH endpoint URL (e.g., "https://dns.google/dns-query").
	//
	// Set by [NewDNSOverHTTPSConnFunc] to the user-provided value.
	URL string

	// ErrClassifier classifies errors for structured logging.
	//
	// Set by [NewDNSOverHTTPSConnFunc] from [Config.ErrClassifier].
	ErrClassifier ErrClassifier

	// Logger is the [SLogger] to use (configurable for testing or custom logging).
	//
	// Set by [NewDNSOverHTTPSConnFunc] to the user-provided logger.
	Logger SLogger

	// TimeNow is the function to get the current time (configurable for testing).
	//
	// Set by [NewDNSOverHTTPSConnFunc] from [Config.TimeNow].
	TimeNow func() time.Time
}

// NewDNSOverHTTPSConnFunc returns a new [*DNSOverHTTPSConnFunc].
//
// The cfg argument contains the common configuration for nop operations.
//
// The url parameter is the DoH endpoint (e.g., "https://dns.google/dns-query").
//
// The logger argument is the [SLogger] to use for structured logging.
func NewDNSOverHTTPSConnFunc(cfg *Config, url string, logger SLogger) *DNSOverHTTPSConnFunc {
	return &DNSOverHTTPSConnFunc{
		URL:           url,
		ErrClassifier: cfg.ErrClassifier,
		Logger:        logger,
		TimeNow:       cfg.TimeNow,
	}
}

var _ Func[*HTTPConn, *DNSOverHTTPSConn] = &DNSOverHTTPSConnFunc{}

// Call wraps the HTTPConn into a DNSOverHTTPSConn.
func (op *DNSOverHTTPSConnFunc) Call(ctx context.Context, httpConn *HTTPConn) (*DNSOverHTTPSConn, error) {
	return &DNSOverHTTPSConn{
		httpConn:      httpConn,
		url:           op.URL,
		ErrClassifier: op.ErrClassifier,
		Logger:        op.Logger,
		TimeNow:       op.TimeNow,
	}, nil
}
