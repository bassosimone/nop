//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Adapted from: https://github.com/rbmk-project/rbmk/blob/v0.17.0/pkg/common/httpslog/httpslog.go
//

package nop

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/bassosimone/safeconn"
	"github.com/bassosimone/sud"
	"golang.org/x/net/http2"
)

// HTTPConn represents an HTTP "connection" (a configured transport over a connection).
//
// The caller is responsible for calling [HTTPConn.Close] when done.
//
// HTTPConn performs round trips with structured logging and transparent body
// observation: httpRoundTripStart/httpRoundTripDone span events are emitted
// around each round trip, and the response body is lazily wrapped to emit
// httpBodyStreamStart/httpBodyStreamDone events.
//
// Construct using [NewHTTPConnFunc], [NewHTTPConnFuncPlain], [NewHTTPConnFuncTLS].
type HTTPConn struct {
	// conn is the underlying connection.
	conn net.Conn

	// txp is the HTTP transport.
	txp http.RoundTripper

	// closeIdleFunc closes idle connections in the transport.
	closeIdleFunc func()

	// ErrClassifier classifies errors for structured logging.
	ErrClassifier ErrClassifier

	// Logger is the [SLogger] to use (configurable for testing or custom logging).
	Logger SLogger

	// TimeNow is the function to get the current time (configurable for testing).
	TimeNow func() time.Time
}

// RoundTrip implements [http.RoundTripper].
func (hc *HTTPConn) RoundTrip(req *http.Request) (*http.Response, error) {
	// 1. Get the underlying connection for logging metadata
	conn := hc.conn

	// 2. Log before the round trip
	t0 := hc.TimeNow()
	deadline, _ := req.Context().Deadline()
	httpLogRoundTripStart(hc, conn, req, t0, deadline)

	// 3. Perform the round trip
	resp, err := hc.txp.RoundTrip(req)

	// 4. Log after the round trip
	httpLogRoundTripDone(hc, conn, req, t0, deadline, resp, err)

	// 5. On error, return immediately
	if err != nil {
		return nil, err
	}

	// 6. Wrap the response body with lazy structured logging
	resp.Body = httpBodyWrap(
		resp.Body,
		hc.ErrClassifier,
		safeconn.LocalAddr(conn),
		hc.Logger,
		safeconn.Network(conn),
		safeconn.RemoteAddr(conn),
		hc.TimeNow,
	)
	return resp, nil
}

// Close cleans up the transport and closes the underlying connection.
func (hc *HTTPConn) Close() error {
	hc.closeIdleFunc()
	return hc.conn.Close()
}

// Conn returns the underlying [net.Conn] used by this [*HTTPConn].
//
// This method exists to support logging operations that need connection
// metadata (local/remote addresses, network type).
func (hc *HTTPConn) Conn() net.Conn {
	return hc.conn
}

func httpLogRoundTripStart(hc *HTTPConn, conn net.Conn, req *http.Request, t0 time.Time, deadline time.Time) {
	hc.Logger.Info(
		"httpRoundTripStart",
		slog.Time("deadline", deadline),
		slog.String("httpMethod", req.Method),
		slog.String("httpUrl", req.URL.String()),
		slog.Any("httpRequestHeaders", req.Header),
		slog.String("localAddr", safeconn.LocalAddr(conn)),
		slog.String("protocol", safeconn.Network(conn)),
		slog.String("remoteAddr", safeconn.RemoteAddr(conn)),
		slog.Time("t", t0),
	)
}

func httpLogRoundTripDone(hc *HTTPConn, conn net.Conn, req *http.Request,
	t0 time.Time, deadline time.Time, resp *http.Response, err error) {
	var (
		statusCode int
		headers    http.Header
	)
	if resp != nil {
		statusCode = resp.StatusCode
		headers = resp.Header
	}
	hc.Logger.Info(
		"httpRoundTripDone",
		slog.Time("deadline", deadline),
		slog.Any("err", err),
		slog.String("errClass", hc.ErrClassifier.Classify(err)),
		slog.String("httpMethod", req.Method),
		slog.String("httpUrl", req.URL.String()),
		slog.Any("httpRequestHeaders", req.Header),
		slog.Any("httpResponseHeaders", headers),
		slog.Int("httpResponseStatusCode", statusCode),
		slog.String("localAddr", safeconn.LocalAddr(conn)),
		slog.String("protocol", safeconn.Network(conn)),
		slog.String("remoteAddr", safeconn.RemoteAddr(conn)),
		slog.Time("t0", t0),
		slog.Time("t", hc.TimeNow()),
	)
}

// HTTPConnFunc wraps a connection into an [*HTTPConn].
//
// This is a generic [Func] that can be composed into pipelines. It creates an
// [*HTTPConn] from the input connection with ALPN-based protocol detection.
//
// Use [HTTPConnFuncPlain] after TCP connect operations for plain HTTP, and use
// [HTTPConnFuncTLS] after TLS handshake operations for HTTPS.
//
// The caller is responsible for closing the returned [*HTTPConn].
//
// All fields are safe to modify after construction but before first use.
// Fields must not be mutated concurrently with calls to [Call].
type HTTPConnFunc[T net.Conn] struct {
	// ErrClassifier classifies errors for structured logging.
	//
	// Set by [NewHTTPConnFunc] from [Config.ErrClassifier].
	ErrClassifier ErrClassifier

	// Logger is the [SLogger] to use (configurable for testing or custom logging).
	//
	// Set by [NewHTTPConnFunc] to the user-provided logger.
	Logger SLogger

	// TimeNow is the function to get the current time (configurable for testing).
	//
	// Set by [NewHTTPConnFunc] from [Config.TimeNow].
	TimeNow func() time.Time
}

// NewHTTPConnFunc returns a new [*HTTPConnFunc].
//
// The cfg argument contains the common configuration for nop operations.
//
// The logger argument is the [SLogger] to use for structured logging.
func NewHTTPConnFunc[T net.Conn](cfg *Config, logger SLogger) *HTTPConnFunc[T] {
	return &HTTPConnFunc[T]{
		ErrClassifier: cfg.ErrClassifier,
		Logger:        logger,
		TimeNow:       cfg.TimeNow,
	}
}

var _ Func[net.Conn, *HTTPConn] = &HTTPConnFunc[net.Conn]{}
var _ Func[TLSConn, *HTTPConn] = &HTTPConnFunc[TLSConn]{}

// Call implements [Func].
func (op *HTTPConnFunc[T]) Call(ctx context.Context, conn T) (*HTTPConn, error) {
	// Obtain the protocol that was negotiated
	type connectionStater interface {
		ConnectionState() tls.ConnectionState
	}
	var alpn string
	if csp, ok := any(conn).(connectionStater); ok {
		alpn = csp.ConnectionState().NegotiatedProtocol
	}

	// Create a special dialer that works just once
	dialer := sud.NewSingleUseDialer(conn)

	// Create proper transport depending on ALPN
	var txp http.RoundTripper
	var closeIdleFunc func()
	switch alpn {
	case "h2":
		h2txp := &http2.Transport{
			DialTLSContext:     dialer.DialTLSContext,
			DisableCompression: false,
		}
		txp = h2txp
		closeIdleFunc = h2txp.CloseIdleConnections

	default:
		h1txp := &http.Transport{
			DialContext:        dialer.DialContext,
			DialTLSContext:     dialer.DialContext,
			DisableKeepAlives:  true,
			DisableCompression: false,
		}
		txp = h1txp
		closeIdleFunc = h1txp.CloseIdleConnections
	}

	hc := &HTTPConn{
		conn:          conn,
		txp:           txp,
		closeIdleFunc: closeIdleFunc,
		ErrClassifier: op.ErrClassifier,
		Logger:        op.Logger,
		TimeNow:       op.TimeNow,
	}
	return hc, nil
}

// NewHTTPConnFuncPlain returns a new [*HTTPConnFunc] for plain HTTP connections.
//
// This is syntactic sugar for NewHTTPConnFunc[net.Conn](cfg, logger).
func NewHTTPConnFuncPlain(cfg *Config, logger SLogger) *HTTPConnFunc[net.Conn] {
	return NewHTTPConnFunc[net.Conn](cfg, logger)
}

// NewHTTPConnFuncTLS returns a new [*HTTPConnFunc] for HTTPS connections.
//
// This is syntactic sugar for NewHTTPConnFunc[TLSConn](cfg, logger).
func NewHTTPConnFuncTLS(cfg *Config, logger SLogger) *HTTPConnFunc[TLSConn] {
	return NewHTTPConnFunc[TLSConn](cfg, logger)
}
