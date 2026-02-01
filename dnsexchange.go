// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"log/slog"
	"time"
)

// dnsExchangeLogContext holds common logging state for DNS exchanges.
//
// This type exists to consolidate the logging boilerplate shared by
// DNS-over-UDP, DNS-over-TCP, and DNS-over-TLS exchange methods.
type dnsExchangeLogContext struct {
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

// logStart logs the start of a DNS exchange.
func (lc *dnsExchangeLogContext) logStart(t0 time.Time, deadline time.Time) {
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

// logDone logs the completion of a DNS exchange.
func (lc *dnsExchangeLogContext) logDone(t0 time.Time, deadline time.Time, err error) {
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

// makeQueryObserver returns an observer function for raw DNS queries.
//
// The rqr pointer is used to capture the raw query for correlation
// with the response observer.
func (lc *dnsExchangeLogContext) makeQueryObserver(t0 time.Time, rqr *[]byte) func([]byte) {
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

// makeResponseObserver returns an observer function for raw DNS responses.
//
// The rqr pointer should be the same one passed to makeQueryObserver,
// allowing the response to be correlated with the original query.
func (lc *dnsExchangeLogContext) makeResponseObserver(t0 time.Time, rqr *[]byte) func([]byte) {
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
