// SPDX-License-Identifier: GPL-3.0-or-later

package nop_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/netip"
	"os"
	"slices"
	"time"

	"github.com/bassosimone/dnscodec"
	"github.com/bassosimone/nop"
	"github.com/bassosimone/runtimex"
	"github.com/miekg/dns"
)

// This example shows how to compose a DNS-over-HTTPS pipeline that
// resolves a domain name using Google's public DNS server.
func Example_dnsOverHTTPS() {
	// Create context with overall timeout for the entire operation.
	// Caller controls timeout externally - nop never modifies the context.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create the shared configuration for nop operations.
	cfg := nop.NewConfig()

	// Create a logger that emits JSON to stderr. Use LevelDebug to include
	// per-I/O events (read, write, deadline); use LevelInfo to see only
	// lifecycle and protocol events (connect, close, TLS, DNS, HTTP).
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Generate a span ID (UUIDv7) and attach it to the logger so that
	// all log entries from this operation can be correlated.
	spanID := nop.NewSpanID()
	logger = logger.With("spanID", spanID)

	// Create pipeline for establishing a DNS-over-HTTPS connection.
	// CancelWatchFunc binds context lifecycle to connection lifecycle:
	// when context is done (timeout, cancel, signal), connection closes.
	epntOp := nop.NewEndpointFunc(netip.MustParseAddrPort("8.8.8.8:443"))

	connectOp := nop.NewConnectFunc(cfg, "tcp", logger)

	observeOp := nop.NewObserveConnFunc(cfg, logger)

	autoCancelOp := nop.NewCancelWatchFunc()

	tlsConfig := &tls.Config{ServerName: "dns.google", NextProtos: []string{"h2", "http/1.1"}}
	tlsHandshakeOp := nop.NewTLSHandshakeFunc(cfg, tlsConfig, logger)

	httpConnOp := nop.NewHTTPConnFuncTLS(cfg, logger)

	wrapOp := nop.NewDNSOverHTTPSConnFunc(cfg, "https://dns.google/dns-query", logger)

	dialPipe := nop.Compose7(epntOp, connectOp, observeOp, autoCancelOp, tlsHandshakeOp, httpConnOp, wrapOp)

	// Connect and wrap in DNSOverHTTPSConn (which owns the underlying HTTPConn)
	dnsConn := runtimex.PanicOnError1(dialPipe.Call(ctx, nop.Unit{}))
	defer dnsConn.Close()

	// Perform the DNS exchange
	dnsQuery := dnscodec.NewQuery("dns.google", dns.TypeA)
	dnsResp := runtimex.PanicOnError1(dnsConn.Exchange(ctx, dnsQuery))

	// Print the results
	addrs := runtimex.PanicOnError1(dnsResp.RecordsA())
	slices.Sort(addrs)
	fmt.Printf("%+v\n", addrs)

	// Output:
	// [8.8.4.4 8.8.8.8]
}
