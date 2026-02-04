// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"log/slog"
	"time"
)

// DNSExchangeLogContext holds common logging state for DNS exchanges.
//
// This type consolidates the logging boilerplate shared by the built-in
// DNS-over-UDP, DNS-over-TCP, DNS-over-TLS, and DNS-over-HTTPS exchange
// methods ([*DNSOverUDPConn], [*DNSOverTCPConn], [*DNSOverTLSConn],
// [*DNSOverHTTPSConn]).
//
// It is also useful for callers that need to implement custom DNS exchange
// loops on top of a raw connection obtained via a nop pipeline. For example,
// a caller collecting duplicate DNS-over-UDP responses for censorship
// detection can use this type to emit structured logs consistent with the
// built-in exchange methods while driving [minest.DNSOverUDPTransport]
// send/receive directly.
type DNSExchangeLogContext struct {
	// ErrClassifier classifies errors for structured logging.
	ErrClassifier ErrClassifier

	// LocalAddr is the local address of the connection.
	LocalAddr string

	// Logger is the SLogger to use.
	Logger SLogger

	// Protocol is the network protocol (e.g., "tcp", "udp").
	Protocol string

	// RemoteAddr is the remote address of the connection.
	RemoteAddr string

	// ServerProtocol is the DNS protocol (e.g., "udp", "tcp", "dot").
	ServerProtocol string

	// TimeNow is the function to get the current time.
	TimeNow func() time.Time
}

// LogStart logs the start of a DNS exchange.
func (lc *DNSExchangeLogContext) LogStart(t0 time.Time, deadline time.Time) {
	lc.Logger.Info(
		"dnsExchangeStart",
		slog.Time("deadline", deadline),
		slog.String("localAddr", lc.LocalAddr),
		slog.String("protocol", lc.Protocol),
		slog.String("remoteAddr", lc.RemoteAddr),
		slog.String("serverProtocol", lc.ServerProtocol),
		slog.Time("t", t0),
	)
}

// LogDone logs the completion of a DNS exchange.
func (lc *DNSExchangeLogContext) LogDone(t0 time.Time, deadline time.Time, err error) {
	lc.Logger.Info(
		"dnsExchangeDone",
		slog.Time("deadline", deadline),
		slog.Any("err", err),
		slog.String("errClass", lc.ErrClassifier.Classify(err)),
		slog.String("localAddr", lc.LocalAddr),
		slog.String("protocol", lc.Protocol),
		slog.String("remoteAddr", lc.RemoteAddr),
		slog.String("serverProtocol", lc.ServerProtocol),
		slog.Time("t0", t0),
		slog.Time("t", lc.TimeNow()),
	)
}

// MakeQueryObserver returns an observer function for raw DNS queries.
//
// The rqr pointer is used to capture the raw query for correlation
// with the response observer.
func (lc *DNSExchangeLogContext) MakeQueryObserver(t0 time.Time, rqr *[]byte) func([]byte) {
	return func(rawQuery []byte) {
		lc.Logger.Info(
			"dnsQuery",
			slog.String("serverProtocol", lc.ServerProtocol),
			slog.Any("dnsRawQuery", rawQuery),
			slog.String("localAddr", lc.LocalAddr),
			slog.String("protocol", lc.Protocol),
			slog.String("remoteAddr", lc.RemoteAddr),
			slog.Time("t", t0),
		)
		*rqr = rawQuery
	}
}

// MakeResponseObserver returns an observer function for raw DNS responses.
//
// The rqr pointer should be the same one passed to [DNSExchangeLogContext.MakeQueryObserver],
// allowing the response to be correlated with the original query.
func (lc *DNSExchangeLogContext) MakeResponseObserver(t0 time.Time, rqr *[]byte) func([]byte) {
	return func(rawResp []byte) {
		lc.Logger.Info(
			"dnsResponse",
			slog.String("serverProtocol", lc.ServerProtocol),
			slog.Any("dnsRawQuery", *rqr),
			slog.String("localAddr", lc.LocalAddr),
			slog.String("protocol", lc.Protocol),
			slog.String("remoteAddr", lc.RemoteAddr),
			slog.Time("t0", t0),
			slog.Time("t", lc.TimeNow()),
			slog.Any("dnsRawResponse", rawResp),
		)
	}
}
