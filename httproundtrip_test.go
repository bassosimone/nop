// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// funcRoundTripper implements http.RoundTripper using a function.
type funcRoundTripper func(*http.Request) (*http.Response, error)

func (f funcRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// RoundTrip delegates to the underlying transport and returns the response.
func TestHTTPConnRoundTripSuccess(t *testing.T) {
	mockConn := newMinimalConn()

	wantResp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/html"}},
		Body:       io.NopCloser(strings.NewReader("OK")),
	}

	httpConn := &HTTPConn{
		conn: mockConn,
		txp: funcRoundTripper(func(req *http.Request) (*http.Response, error) {
			return wantResp, nil
		}),
		closeIdleFunc: func() {},
		ErrClassifier: NewConfig().ErrClassifier,
		Logger:        DefaultSLogger(),
		TimeNow:       time.Now,
	}

	req, err := http.NewRequest("GET", "https://example.com/", nil)
	require.NoError(t, err)

	resp, err := httpConn.RoundTrip(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "text/html", resp.Header.Get("Content-Type"))
}

// RoundTrip propagates errors from the underlying transport.
func TestHTTPConnRoundTripError(t *testing.T) {
	wantErr := errors.New("round trip failed")

	mockConn := newMinimalConn()

	httpConn := &HTTPConn{
		conn: mockConn,
		txp: funcRoundTripper(func(req *http.Request) (*http.Response, error) {
			return nil, wantErr
		}),
		closeIdleFunc: func() {},
		ErrClassifier: NewConfig().ErrClassifier,
		Logger:        DefaultSLogger(),
		TimeNow:       time.Now,
	}

	req, err := http.NewRequest("GET", "https://example.com/", nil)
	require.NoError(t, err)

	resp, err := httpConn.RoundTrip(req)

	require.ErrorIs(t, err, wantErr)
	assert.Nil(t, resp)
}

// RoundTrip propagates the caller's context deadline to the transport.
func TestHTTPConnRoundTripCallerTimeout(t *testing.T) {
	// Caller-provided timeout
	callerTimeout := 5 * time.Second

	mockConn := newMinimalConn()

	httpConn := &HTTPConn{
		conn: mockConn,
		txp: funcRoundTripper(func(req *http.Request) (*http.Response, error) {
			// Verify context has the caller-provided deadline
			deadline, ok := req.Context().Deadline()
			assert.True(t, ok, "context should have deadline from caller")
			assert.True(t, time.Until(deadline) <= callerTimeout)
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		}),
		closeIdleFunc: func() {},
		ErrClassifier: NewConfig().ErrClassifier,
		Logger:        DefaultSLogger(),
		TimeNow:       time.Now,
	}

	req, err := http.NewRequest("GET", "https://example.com/", nil)
	require.NoError(t, err)

	// Caller provides timeout via context
	ctx, cancel := context.WithTimeout(context.Background(), callerTimeout)
	defer cancel()
	req = req.WithContext(ctx)

	_, err = httpConn.RoundTrip(req)
	require.NoError(t, err)
}

// RoundTrip emits httpRoundTripStart/httpRoundTripDone log events.
func TestHTTPConnRoundTripLogging(t *testing.T) {
	logger, records := newCapturingLogger()

	mockConn := newMinimalConn()

	httpConn := &HTTPConn{
		conn: mockConn,
		txp: funcRoundTripper(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		}),
		closeIdleFunc: func() {},
		ErrClassifier: NewConfig().ErrClassifier,
		Logger:        logger,
		TimeNow:       time.Now,
	}

	req, err := http.NewRequest("GET", "https://example.com/", nil)
	require.NoError(t, err)

	_, _ = httpConn.RoundTrip(req)

	require.Len(t, *records, 2)
	assert.Equal(t, "httpRoundTripStart", (*records)[0].Message)
	assert.Equal(t, "httpRoundTripDone", (*records)[1].Message)
}

// RoundTrip logs localAddr, remoteAddr, and protocol in the done event.
func TestHTTPConnRoundTripLogsConnectionMetadata(t *testing.T) {
	wantLocalAddr := "127.0.0.1:54321"
	wantRemoteAddr := "93.184.216.34:443"
	wantProtocol := "tcp"

	logger, records := newCapturingLogger()

	mockConn := newMinimalConn()
	mockConn.LocalAddrFunc = func() net.Addr {
		return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 54321}
	}
	mockConn.RemoteAddrFunc = func() net.Addr {
		return &net.TCPAddr{IP: net.IPv4(93, 184, 216, 34), Port: 443}
	}

	httpConn := &HTTPConn{
		conn: mockConn,
		txp: funcRoundTripper(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		}),
		closeIdleFunc: func() {},
		ErrClassifier: NewConfig().ErrClassifier,
		Logger:        logger,
		TimeNow:       time.Now,
	}

	req, err := http.NewRequest("GET", "https://example.com/", nil)
	require.NoError(t, err)

	_, err = httpConn.RoundTrip(req)
	require.NoError(t, err)

	// Check the httpRoundTripDone record for connection metadata attributes
	require.Len(t, *records, 2)
	doneRecord := (*records)[1]

	// Extract attributes from the log record
	var gotLocalAddr, gotRemoteAddr, gotProtocol string
	doneRecord.Attrs(func(attr slog.Attr) bool {
		switch attr.Key {
		case "localAddr":
			gotLocalAddr = attr.Value.String()
		case "remoteAddr":
			gotRemoteAddr = attr.Value.String()
		case "protocol":
			gotProtocol = attr.Value.String()
		}
		return true
	})

	assert.Equal(t, wantLocalAddr, gotLocalAddr)
	assert.Equal(t, wantRemoteAddr, gotRemoteAddr)
	assert.Equal(t, wantProtocol, gotProtocol)
}
