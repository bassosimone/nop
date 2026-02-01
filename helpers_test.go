// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net"

	"github.com/bassosimone/netstub"
	"github.com/bassosimone/slogstub"
	"github.com/bassosimone/tlsstub"
)

// newCapturingLogger returns a logger that captures all log records into the
// returned slice. The caller can inspect the slice after exercising the code
// under test to verify which events were emitted.
func newCapturingLogger() (*slog.Logger, *[]slog.Record) {
	var records []slog.Record
	handler := &slogstub.FuncHandler{
		EnabledFunc: func(ctx context.Context, level slog.Level) bool {
			return true
		},
		HandleFunc: func(ctx context.Context, record slog.Record) error {
			records = append(records, record)
			return nil
		},
	}
	return slog.New(handler), &records
}

// newMockTLSEngine returns a [*tlsstub.FuncTLSEngine] that wraps the given
// [TLSConn]. The engine's ClientFunc returns the conn, NameFunc returns
// "mock", and ParrotFunc returns "".
func newMockTLSEngine(conn TLSConn) *tlsstub.FuncTLSEngine[TLSConn] {
	return &tlsstub.FuncTLSEngine[TLSConn]{
		ClientFunc: func(c net.Conn, config *tls.Config) TLSConn {
			return conn
		},
		NameFunc: func() string {
			return "mock"
		},
		ParrotFunc: func() string {
			return ""
		},
	}
}

// newMinimalConn returns a [*netstub.FuncConn] with only LocalAddrFunc and
// RemoteAddrFunc set. This is the minimum needed for code that calls
// [safeconn.LocalAddr], [safeconn.RemoteAddr], and [safeconn.Network]
// during construction.
func newMinimalConn() *netstub.FuncConn {
	return &netstub.FuncConn{
		LocalAddrFunc:  func() net.Addr { return &net.TCPAddr{} },
		RemoteAddrFunc: func() net.Addr { return &net.TCPAddr{} },
	}
}
