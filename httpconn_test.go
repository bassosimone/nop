// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"testing"

	"github.com/bassosimone/tlsstub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Call wraps the connection in an HTTP transport and selects HTTP/1.1 or HTTP/2 based on ALPN.
func TestNewHTTPConn(t *testing.T) {
	t.Run("plain connection uses HTTP/1.1", func(t *testing.T) {
		mockConn := newMinimalConn()

		fn := NewHTTPConnFuncPlain(NewConfig(), DefaultSLogger())
		hc, err := fn.Call(context.Background(), mockConn)
		require.NoError(t, err)

		require.NotNil(t, hc)
		assert.NotNil(t, hc.Conn())
		assert.Equal(t, mockConn, hc.Conn())
	})

	t.Run("TLS connection with h2 ALPN uses HTTP/2", func(t *testing.T) {
		mockConn := &tlsstub.FuncTLSConn{
			FuncConn: newMinimalConn(),
			ConnectionStateFunc: func() tls.ConnectionState {
				return tls.ConnectionState{NegotiatedProtocol: "h2"}
			},
			HandshakeContextFunc: func(ctx context.Context) error {
				return nil
			},
		}

		fn := NewHTTPConnFuncTLS(NewConfig(), DefaultSLogger())
		hc, err := fn.Call(context.Background(), mockConn)
		require.NoError(t, err)

		require.NotNil(t, hc)
		assert.NotNil(t, hc.Conn())
	})

	t.Run("TLS connection without ALPN uses HTTP/1.1", func(t *testing.T) {
		mockConn := &tlsstub.FuncTLSConn{
			FuncConn: newMinimalConn(),
			ConnectionStateFunc: func() tls.ConnectionState {
				return tls.ConnectionState{NegotiatedProtocol: ""}
			},
			HandshakeContextFunc: func(ctx context.Context) error {
				return nil
			},
		}

		fn := NewHTTPConnFuncTLS(NewConfig(), DefaultSLogger())
		hc, err := fn.Call(context.Background(), mockConn)
		require.NoError(t, err)

		require.NotNil(t, hc)
	})
}

// Close delegates to the underlying connection.
func TestHTTPConnClose(t *testing.T) {
	closeCalled := false
	mockConn := newMinimalConn()
	mockConn.CloseFunc = func() error {
		closeCalled = true
		return nil
	}

	fn := NewHTTPConnFuncPlain(NewConfig(), DefaultSLogger())
	hc, err := fn.Call(context.Background(), mockConn)
	require.NoError(t, err)

	err = hc.Close()

	require.NoError(t, err)
	assert.True(t, closeCalled)
}

// Close propagates errors from the underlying connection.
func TestHTTPConnCloseError(t *testing.T) {
	wantErr := errors.New("close error")

	mockConn := newMinimalConn()
	mockConn.CloseFunc = func() error {
		return wantErr
	}

	fn := NewHTTPConnFuncPlain(NewConfig(), DefaultSLogger())
	hc, err := fn.Call(context.Background(), mockConn)
	require.NoError(t, err)

	err = hc.Close()

	require.ErrorIs(t, err, wantErr)
}

// Conn returns the underlying net.Conn.
func TestHTTPConnConn(t *testing.T) {
	mockConn := newMinimalConn()

	fn := NewHTTPConnFuncPlain(NewConfig(), DefaultSLogger())
	hc, err := fn.Call(context.Background(), mockConn)
	require.NoError(t, err)

	assert.Equal(t, mockConn, hc.Conn())
}

// NewHTTPConnFuncPlain satisfies Func[net.Conn, *HTTPConn].
func TestNewHTTPConnFuncPlain(t *testing.T) {
	fn := NewHTTPConnFuncPlain(NewConfig(), DefaultSLogger())
	require.NotNil(t, fn)

	// Verify it satisfies Func interface
	var _ Func[net.Conn, *HTTPConn] = fn
}

// NewHTTPConnFuncTLS satisfies Func[TLSConn, *HTTPConn].
func TestNewHTTPConnFuncTLS(t *testing.T) {
	fn := NewHTTPConnFuncTLS(NewConfig(), DefaultSLogger())
	require.NotNil(t, fn)

	// Verify it satisfies Func interface
	var _ Func[TLSConn, *HTTPConn] = fn
}
