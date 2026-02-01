// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/bassosimone/dnscodec"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NewDNSOverHTTPSConnFunc populates all fields from Config and the provided logger.
func TestNewDNSOverHTTPSConnFunc(t *testing.T) {
	cfg := NewConfig()
	url := "https://dns.google/dns-query"
	logger := DefaultSLogger()

	fn := NewDNSOverHTTPSConnFunc(cfg, url, logger)

	require.NotNil(t, fn)
	assert.Equal(t, url, fn.URL)
	assert.NotNil(t, fn.Logger)
	assert.NotNil(t, fn.TimeNow)
	assert.NotNil(t, fn.ErrClassifier)
}

// Call wraps the HTTPConn and populates all observable fields.
func TestDNSOverHTTPSConnFuncCall(t *testing.T) {
	cfg := NewConfig()
	url := "https://dns.google/dns-query"

	mockConn := newMinimalConn()

	httpConnFunc := NewHTTPConnFuncPlain(cfg, DefaultSLogger())
	httpConn, err := httpConnFunc.Call(context.Background(), mockConn)
	require.NoError(t, err)

	fn := NewDNSOverHTTPSConnFunc(cfg, url, DefaultSLogger())
	result, err := fn.Call(context.Background(), httpConn)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify the conn is wrapped correctly
	assert.Equal(t, httpConn, result.HTTPConn())
	assert.NotNil(t, result.Logger)
	assert.NotNil(t, result.TimeNow)
	assert.NotNil(t, result.ErrClassifier)
}

// Close delegates to the underlying HTTPConn.
func TestDNSOverHTTPSConnClose(t *testing.T) {
	closeCalled := false
	mockConn := newMinimalConn()
	mockConn.CloseFunc = func() error {
		closeCalled = true
		return nil
	}

	cfg := NewConfig()
	httpConnFunc := NewHTTPConnFuncPlain(cfg, DefaultSLogger())
	httpConn, err := httpConnFunc.Call(context.Background(), mockConn)
	require.NoError(t, err)

	fn := NewDNSOverHTTPSConnFunc(cfg, "https://dns.google/dns-query", DefaultSLogger())
	result, err := fn.Call(context.Background(), httpConn)
	require.NoError(t, err)

	err = result.Close()

	require.NoError(t, err)
	assert.True(t, closeCalled)
}

// HTTPConn returns the underlying *HTTPConn.
func TestDNSOverHTTPSConnHTTPConn(t *testing.T) {
	mockConn := newMinimalConn()

	cfg := NewConfig()
	httpConnFunc := NewHTTPConnFuncPlain(cfg, DefaultSLogger())
	httpConn, err := httpConnFunc.Call(context.Background(), mockConn)
	require.NoError(t, err)

	fn := NewDNSOverHTTPSConnFunc(cfg, "https://dns.google/dns-query", DefaultSLogger())
	result, err := fn.Call(context.Background(), httpConn)
	require.NoError(t, err)

	assert.Equal(t, httpConn, result.HTTPConn())
}

// Exchange propagates errors from the HTTP round trip.
func TestDNSOverHTTPSConnExchangeRoundTripError(t *testing.T) {
	wantErr := errors.New("round trip error")

	httpConn := &HTTPConn{
		conn: newMinimalConn(),
		txp: funcRoundTripper(func(req *http.Request) (*http.Response, error) {
			return nil, wantErr
		}),
		closeIdleFunc: func() {},
		ErrClassifier: NewConfig().ErrClassifier,
		Logger:        DefaultSLogger(),
		TimeNow:       time.Now,
	}

	cfg := NewConfig()
	fn := NewDNSOverHTTPSConnFunc(cfg, "https://dns.google/dns-query", DefaultSLogger())
	result, err := fn.Call(context.Background(), httpConn)
	require.NoError(t, err)

	query := dnscodec.NewQuery("example.com", dns.TypeA)
	_, err = result.Exchange(context.Background(), query)

	require.Error(t, err)
}

// Exchange returns an error when the URL is invalid.
func TestDNSOverHTTPSConnExchangeInvalidURL(t *testing.T) {
	mockConn := newMinimalConn()

	cfg := NewConfig()
	httpConnFunc := NewHTTPConnFuncPlain(cfg, DefaultSLogger())
	httpConn, err := httpConnFunc.Call(context.Background(), mockConn)
	require.NoError(t, err)

	fn := NewDNSOverHTTPSConnFunc(cfg, "\t", DefaultSLogger())
	result, err := fn.Call(context.Background(), httpConn)
	require.NoError(t, err)

	query := dnscodec.NewQuery("example.com", dns.TypeA)
	_, err = result.Exchange(context.Background(), query)

	require.Error(t, err)
}
