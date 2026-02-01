// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/bassosimone/netstub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NewConnectFunc populates all fields from Config and the provided logger.
func TestNewConnectFunc(t *testing.T) {
	cfg := NewConfig()
	logger := DefaultSLogger()

	fn := NewConnectFunc(cfg, "tcp", logger)

	require.NotNil(t, fn)
	assert.Equal(t, "tcp", fn.Network)
	assert.NotNil(t, fn.Dialer)
	assert.NotNil(t, fn.Logger)
	assert.NotNil(t, fn.TimeNow)
	assert.NotNil(t, fn.ErrClassifier)
}

// Call dials the address and returns a net.Conn or an error.
func TestConnectFunc(t *testing.T) {
	tests := []struct {
		// name describes what this test case verifies.
		name string

		// dialer is the mock dialer to use.
		dialer *netstub.FuncDialer

		// network is the network type.
		network string

		// address is the target address.
		address netip.AddrPort

		// wantErr indicates whether we expect an error.
		wantErr bool
	}{
		{
			name: "successful TCP connect",
			dialer: &netstub.FuncDialer{
				DialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
					conn := newMinimalConn()
					conn.CloseFunc = func() error { return nil }
					conn.LocalAddrFunc = func() net.Addr {
						return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 54321}
					}
					conn.RemoteAddrFunc = func() net.Addr {
						return &net.TCPAddr{IP: net.IPv4(93, 184, 216, 34), Port: 443}
					}
					return conn, nil
				},
			},
			network: "tcp",
			address: netip.MustParseAddrPort("93.184.216.34:443"),
			wantErr: false,
		},

		{
			name: "dial error",
			dialer: &netstub.FuncDialer{
				DialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
					return nil, errors.New("connection refused")
				},
			},
			network: "tcp",
			address: netip.MustParseAddrPort("93.184.216.34:443"),
			wantErr: true,
		},

		{
			name: "successful UDP connect",
			dialer: &netstub.FuncDialer{
				DialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
					conn := newMinimalConn()
					conn.CloseFunc = func() error { return nil }
					conn.LocalAddrFunc = func() net.Addr {
						return &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 54321}
					}
					conn.RemoteAddrFunc = func() net.Addr {
						return &net.UDPAddr{IP: net.IPv4(8, 8, 8, 8), Port: 53}
					}
					return conn, nil
				},
			},
			network: "udp",
			address: netip.MustParseAddrPort("8.8.8.8:53"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig()
			cfg.Dialer = tt.dialer

			fn := NewConnectFunc(cfg, tt.network, DefaultSLogger())
			conn, err := fn.Call(context.Background(), tt.address)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, conn)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, conn)
			conn.Close()
		})
	}
}

// Call transparently passes the caller's context to the dialer.
func TestConnectFuncContextTransparency(t *testing.T) {
	tests := []struct {
		// name describes the scenario.
		name string

		// dialer is the mock dialer to use.
		dialer *netstub.FuncDialer

		// makeCtx builds the context for the call.
		makeCtx func() (context.Context, context.CancelFunc)
	}{
		{
			name: "pre-expired context",
			dialer: &netstub.FuncDialer{
				DialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
					if ctx.Err() != nil {
						return nil, ctx.Err()
					}
					return nil, errors.New("should not reach here")
				},
			},
			makeCtx: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
				time.Sleep(10 * time.Millisecond)
				return ctx, cancel
			},
		},

		{
			name: "context expires during dial",
			dialer: &netstub.FuncDialer{
				DialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
					time.Sleep(10 * time.Millisecond)
					if ctx.Err() != nil {
						return nil, ctx.Err()
					}
					return nil, errors.New("should not reach here")
				},
			},
			makeCtx: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 1*time.Nanosecond)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig()
			cfg.Dialer = tt.dialer

			fn := NewConnectFunc(cfg, "tcp", DefaultSLogger())

			ctx, cancel := tt.makeCtx()
			defer cancel()

			_, err := fn.Call(ctx, netip.MustParseAddrPort("93.184.216.34:443"))
			require.Error(t, err)
		})
	}
}

// Call propagates the caller's context deadline to the dialer.
func TestConnectFuncCallerContextDeadline(t *testing.T) {
	cfg := NewConfig()
	dialCalled := false
	expectedTimeout := 5 * time.Second
	cfg.Dialer = &netstub.FuncDialer{
		DialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
			dialCalled = true
			deadline, ok := ctx.Deadline()
			assert.True(t, ok, "context should have deadline from caller")
			assert.True(t, time.Until(deadline) <= expectedTimeout)
			return nil, errors.New("expected error")
		},
	}

	fn := NewConnectFunc(cfg, "tcp", DefaultSLogger())

	// Caller controls timeout via context.WithTimeout
	ctx, cancel := context.WithTimeout(context.Background(), expectedTimeout)
	defer cancel()

	_, _ = fn.Call(ctx, netip.MustParseAddrPort("93.184.216.34:443"))

	assert.True(t, dialCalled)
}

// Call emits connectStart/connectDone log events.
func TestConnectFuncLogging(t *testing.T) {
	logger, records := newCapturingLogger()

	cfg := NewConfig()
	cfg.Dialer = &netstub.FuncDialer{
		DialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
			conn := newMinimalConn()
			conn.CloseFunc = func() error { return nil }
			return conn, nil
		},
	}

	fn := NewConnectFunc(cfg, "tcp", logger)
	conn, err := fn.Call(context.Background(), netip.MustParseAddrPort("93.184.216.34:443"))
	require.NoError(t, err)
	conn.Close()

	require.Len(t, *records, 2)
	assert.Equal(t, "connectStart", (*records)[0].Message)
	assert.Equal(t, "connectDone", (*records)[1].Message)
}
