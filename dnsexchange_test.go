// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestLogContext returns a dnsExchangeLogContext wired to a capturing
// logger with fixed metadata, suitable for verifying log output.
func newTestLogContext(logger SLogger) *dnsExchangeLogContext {
	return &dnsExchangeLogContext{
		ErrClassifier:  DefaultErrClassifier,
		LocalAddr:      "127.0.0.1:54321",
		Logger:         logger,
		Protocol:       "udp",
		RemoteAddr:     "8.8.8.8:53",
		ServerProtocol: "udp",
		TimeNow:        time.Now,
	}
}

// logStart emits a dnsExchangeStart event.
func TestDNSExchangeLogContextLogStart(t *testing.T) {
	logger, records := newCapturingLogger()
	lc := newTestLogContext(logger)

	t0 := time.Now()
	deadline := t0.Add(5 * time.Second)
	lc.logStart(t0, deadline)

	require.Len(t, *records, 1)
	assert.Equal(t, "dnsExchangeStart", (*records)[0].Message)
}

// logDone emits a dnsExchangeDone event with error classification.
func TestDNSExchangeLogContextLogDone(t *testing.T) {
	logger, records := newCapturingLogger()
	lc := newTestLogContext(logger)

	t0 := time.Now()
	deadline := t0.Add(5 * time.Second)
	lc.logDone(t0, deadline, nil)

	require.Len(t, *records, 1)
	assert.Equal(t, "dnsExchangeDone", (*records)[0].Message)
}

// logDone includes the error when one is provided.
func TestDNSExchangeLogContextLogDoneWithError(t *testing.T) {
	logger, records := newCapturingLogger()
	lc := newTestLogContext(logger)

	t0 := time.Now()
	deadline := t0.Add(5 * time.Second)
	wantErr := errors.New("timeout")
	lc.logDone(t0, deadline, wantErr)

	require.Len(t, *records, 1)
	assert.Equal(t, "dnsExchangeDone", (*records)[0].Message)

	var gotErr error
	(*records)[0].Attrs(func(attr slog.Attr) bool {
		if attr.Key == "err" {
			gotErr, _ = attr.Value.Any().(error)
			return false
		}
		return true
	})
	assert.Equal(t, wantErr, gotErr)
}

// makeQueryObserver returns a function that emits a dnsQuery event
// and captures the raw query bytes into the provided pointer.
func TestDNSExchangeLogContextMakeQueryObserver(t *testing.T) {
	logger, records := newCapturingLogger()
	lc := newTestLogContext(logger)

	var rqr []byte
	t0 := time.Now()
	observer := lc.makeQueryObserver(t0, &rqr)

	rawQuery := []byte{0x00, 0x01, 0x02}
	observer(rawQuery)

	require.Len(t, *records, 1)
	assert.Equal(t, "dnsQuery", (*records)[0].Message)
	assert.Equal(t, rawQuery, rqr, "raw query should be captured")
}

// makeResponseObserver returns a function that emits a dnsResponse event
// and includes the previously-captured raw query for correlation.
func TestDNSExchangeLogContextMakeResponseObserver(t *testing.T) {
	logger, records := newCapturingLogger()
	lc := newTestLogContext(logger)

	// Simulate the query observer having captured the raw query
	rawQuery := []byte{0x00, 0x01, 0x02}
	rqr := rawQuery

	t0 := time.Now()
	observer := lc.makeResponseObserver(t0, &rqr)

	rawResp := []byte{0x03, 0x04, 0x05}
	observer(rawResp)

	require.Len(t, *records, 1)
	assert.Equal(t, "dnsResponse", (*records)[0].Message)

	// Verify both raw query and response are present in the log
	var gotQuery, gotResp []byte
	(*records)[0].Attrs(func(attr slog.Attr) bool {
		switch attr.Key {
		case "dnsRawQuery":
			gotQuery, _ = attr.Value.Any().([]byte)
		case "dnsRawResponse":
			gotResp, _ = attr.Value.Any().([]byte)
		}
		return true
	})
	assert.Equal(t, rawQuery, gotQuery)
	assert.Equal(t, rawResp, gotResp)
}
