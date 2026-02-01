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

// This example shows how to compose a DNS-over-TLS pipeline that
// resolves a domain name using Google's public DNS server.
func Example_dnsOverTLS() {
	// Create context with overall timeout for the entire operation.
	// Caller controls timeout externally - nop never modifies the context.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a config and logger with a span ID for correlating log entries
	cfg := nop.NewConfig()
	spanID := nop.NewSpanID()
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil)).With("spanID", spanID)

	// Create pipeline for establishing a DNS-over-TLS connection.
	// CancelWatchFunc binds context lifecycle to connection lifecycle:
	// when context is done (timeout, cancel, signal), connection closes.
	epntOp := nop.NewEndpointFunc(netip.MustParseAddrPort("8.8.8.8:853"))

	connectOp := nop.NewConnectFunc(cfg, "tcp", logger)

	observeOp := nop.NewObserveConnFunc(cfg, logger)

	autoCancelOp := nop.NewCancelWatchFunc()

	tlsConfig := &tls.Config{ServerName: "dns.google", NextProtos: []string{"dot"}}
	tlsHandshakeOp := nop.NewTLSHandshakeFunc(cfg, tlsConfig, logger)

	wrapOp := nop.NewDNSOverTLSConnFunc(cfg, logger)

	dialPipe := nop.Compose6(epntOp, connectOp, observeOp, autoCancelOp, tlsHandshakeOp, wrapOp)

	// Connect and wrap in DNSOverTLSConn (which owns the underlying connection)
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
