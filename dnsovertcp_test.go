// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"context"
	"errors"
	"testing"

	"github.com/bassosimone/dnscodec"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NewDNSOverTCPConnFunc populates all fields from Config and the provided logger.
func TestNewDNSOverTCPConnFunc(t *testing.T) {
	cfg := NewConfig()
	logger := DefaultSLogger()

	fn := NewDNSOverTCPConnFunc(cfg, logger)

	require.NotNil(t, fn)
	assert.NotNil(t, fn.Logger)
	assert.NotNil(t, fn.TimeNow)
	assert.NotNil(t, fn.ErrClassifier)
}

// Call wraps the connection and populates all observable fields.
func TestDNSOverTCPConnFuncCall(t *testing.T) {
	cfg := NewConfig()

	mockConn := newMinimalConn()

	fn := NewDNSOverTCPConnFunc(cfg, DefaultSLogger())
	result, err := fn.Call(context.Background(), mockConn)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify the conn is wrapped correctly
	assert.Equal(t, mockConn, result.Conn())
	assert.NotNil(t, result.Logger)
	assert.NotNil(t, result.TimeNow)
	assert.NotNil(t, result.ErrClassifier)
}

// Close delegates to the underlying connection.
func TestDNSOverTCPConnClose(t *testing.T) {
	closeCalled := false
	mockConn := newMinimalConn()
	mockConn.CloseFunc = func() error {
		closeCalled = true
		return nil
	}

	cfg := NewConfig()
	fn := NewDNSOverTCPConnFunc(cfg, DefaultSLogger())
	result, _ := fn.Call(context.Background(), mockConn)

	err := result.Close()

	require.NoError(t, err)
	assert.True(t, closeCalled)
}

// Conn returns the underlying net.Conn.
func TestDNSOverTCPConnConn(t *testing.T) {
	mockConn := newMinimalConn()

	cfg := NewConfig()
	fn := NewDNSOverTCPConnFunc(cfg, DefaultSLogger())
	result, _ := fn.Call(context.Background(), mockConn)

	assert.Equal(t, mockConn, result.Conn())
}

// Exchange propagates write errors from the underlying connection.
func TestDNSOverTCPConnExchangeWriteError(t *testing.T) {
	wantErr := errors.New("write error")

	mockConn := newMinimalConn()
	mockConn.WriteFunc = func(b []byte) (int, error) {
		return 0, wantErr
	}

	cfg := NewConfig()
	fn := NewDNSOverTCPConnFunc(cfg, DefaultSLogger())
	result, err := fn.Call(context.Background(), mockConn)
	require.NoError(t, err)

	query := dnscodec.NewQuery("example.com", dns.TypeA)
	_, err = result.Exchange(context.Background(), query)

	require.Error(t, err)
}
