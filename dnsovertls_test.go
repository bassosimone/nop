// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"context"
	"crypto/tls"
	"errors"
	"testing"

	"github.com/bassosimone/dnscodec"
	"github.com/bassosimone/tlsstub"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NewDNSOverTLSConnFunc populates all fields from Config and the provided logger.
func TestNewDNSOverTLSConnFunc(t *testing.T) {
	cfg := NewConfig()
	logger := DefaultSLogger()

	fn := NewDNSOverTLSConnFunc(cfg, logger)

	require.NotNil(t, fn)
	assert.NotNil(t, fn.Logger)
	assert.NotNil(t, fn.TimeNow)
	assert.NotNil(t, fn.ErrClassifier)
}

// Call wraps the TLS connection and populates all observable fields.
func TestDNSOverTLSConnFuncCall(t *testing.T) {
	cfg := NewConfig()

	mockTLSConn := &tlsstub.FuncTLSConn{
		FuncConn: newMinimalConn(),
		ConnectionStateFunc: func() tls.ConnectionState {
			return tls.ConnectionState{}
		},
		HandshakeContextFunc: func(ctx context.Context) error {
			return nil
		},
	}

	fn := NewDNSOverTLSConnFunc(cfg, DefaultSLogger())
	result, err := fn.Call(context.Background(), mockTLSConn)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify the conn is wrapped correctly
	assert.Equal(t, mockTLSConn, result.Conn())
	assert.NotNil(t, result.Logger)
	assert.NotNil(t, result.TimeNow)
	assert.NotNil(t, result.ErrClassifier)
}

// Close delegates to the underlying TLS connection.
func TestDNSOverTLSConnClose(t *testing.T) {
	closeCalled := false
	mockTLSConn := &tlsstub.FuncTLSConn{
		FuncConn: newMinimalConn(),
		ConnectionStateFunc: func() tls.ConnectionState {
			return tls.ConnectionState{}
		},
		HandshakeContextFunc: func(ctx context.Context) error {
			return nil
		},
	}
	mockTLSConn.FuncConn.CloseFunc = func() error {
		closeCalled = true
		return nil
	}

	cfg := NewConfig()
	fn := NewDNSOverTLSConnFunc(cfg, DefaultSLogger())
	result, _ := fn.Call(context.Background(), mockTLSConn)

	err := result.Close()

	require.NoError(t, err)
	assert.True(t, closeCalled)
}

// Conn returns the underlying TLSConn.
func TestDNSOverTLSConnConn(t *testing.T) {
	mockTLSConn := &tlsstub.FuncTLSConn{
		FuncConn: newMinimalConn(),
		ConnectionStateFunc: func() tls.ConnectionState {
			return tls.ConnectionState{}
		},
		HandshakeContextFunc: func(ctx context.Context) error {
			return nil
		},
	}

	cfg := NewConfig()
	fn := NewDNSOverTLSConnFunc(cfg, DefaultSLogger())
	result, _ := fn.Call(context.Background(), mockTLSConn)

	assert.Equal(t, mockTLSConn, result.Conn())
}

// Exchange propagates write errors from the underlying TLS connection.
func TestDNSOverTLSConnExchangeWriteError(t *testing.T) {
	wantErr := errors.New("write error")

	mockTLSConn := &tlsstub.FuncTLSConn{
		FuncConn: newMinimalConn(),
		ConnectionStateFunc: func() tls.ConnectionState {
			return tls.ConnectionState{}
		},
		HandshakeContextFunc: func(ctx context.Context) error {
			return nil
		},
	}
	mockTLSConn.FuncConn.WriteFunc = func(b []byte) (int, error) {
		return 0, wantErr
	}

	cfg := NewConfig()
	fn := NewDNSOverTLSConnFunc(cfg, DefaultSLogger())
	result, err := fn.Call(context.Background(), mockTLSConn)
	require.NoError(t, err)

	query := dnscodec.NewQuery("example.com", dns.TypeA)
	_, err = result.Exchange(context.Background(), query)

	require.Error(t, err)
}
