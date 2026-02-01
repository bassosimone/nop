// SPDX-License-Identifier: GPL-3.0-or-later

package nop_test

import (
	"context"
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

// This example shows how to compose a DNS-over-UDP pipeline that
// resolves a domain name using Google's public DNS server.
func Example_dnsOverUDP() {
	// Create context with overall timeout for the entire operation.
	// Caller controls timeout externally - nop never modifies the context.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a config and logger with a span ID for correlating log entries
	cfg := nop.NewConfig()
	spanID := nop.NewSpanID()
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil)).With("spanID", spanID)

	// Create pipeline for establishing a DNS-over-UDP connection.
	// CancelWatchFunc binds context lifecycle to connection lifecycle:
	// when context is done (timeout, cancel, signal), connection closes.
	epntOp := nop.NewEndpointFunc(netip.MustParseAddrPort("8.8.8.8:53"))

	connectOp := nop.NewConnectFunc(cfg, "udp", logger)

	observeOp := nop.NewObserveConnFunc(cfg, logger)

	autoCancelOp := nop.NewCancelWatchFunc()

	wrapOp := nop.NewDNSOverUDPConnFunc(cfg, logger)

	dialPipe := nop.Compose5(epntOp, connectOp, observeOp, autoCancelOp, wrapOp)

	// Connect and wrap in DNSOverUDPConn (which owns the underlying connection)
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
