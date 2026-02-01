// SPDX-License-Identifier: GPL-3.0-or-later

package nop_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/netip"
	"os"
	"strings"
	"time"

	"github.com/bassosimone/nop"
	"github.com/bassosimone/runtimex"
)

// This example shows how to compose an HTTPS pipeline that performs
// an HTTP round trip and reads the response body.
func Example_httpsRoundTrip() {
	// Create context with overall timeout for the entire operation.
	// Caller controls timeout externally - nop never modifies the context.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a config and logger with a span ID for correlating log entries
	cfg := nop.NewConfig()
	spanID := nop.NewSpanID()
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil)).With("spanID", spanID)

	// Create pipeline for establishing an HTTPS connection.
	// CancelWatchFunc binds context lifecycle to connection lifecycle:
	// when context is done (timeout, cancel, signal), connection closes.
	epntOp := nop.NewEndpointFunc(netip.MustParseAddrPort("8.8.8.8:443"))

	connectOp := nop.NewConnectFunc(cfg, "tcp", logger)

	observeOp := nop.NewObserveConnFunc(cfg, logger)

	autoCancelOp := nop.NewCancelWatchFunc()

	tlsConfig := &tls.Config{ServerName: "dns.google", NextProtos: []string{"h2", "http/1.1"}}
	tlsHandshakeOp := nop.NewTLSHandshakeFunc(cfg, tlsConfig, logger)

	httpConnOp := nop.NewHTTPConnFuncTLS(cfg, logger)

	dialPipe := nop.Compose6(epntOp, connectOp, observeOp, autoCancelOp, tlsHandshakeOp, httpConnOp)

	// Connect and wrap in HTTPConn
	httpConn := runtimex.PanicOnError1(dialPipe.Call(ctx, nop.Unit{}))
	defer httpConn.Close()

	// Create the HTTP request and perform the round trip
	httpReq := runtimex.PanicOnError1(
		http.NewRequestWithContext(ctx, "GET", "https://dns.google/", http.NoBody))
	resp := runtimex.PanicOnError1(httpConn.RoundTrip(httpReq))
	defer resp.Body.Close()
	runtimex.Assert(resp.StatusCode < 400)

	// Read the body
	body := runtimex.PanicOnError1(io.ReadAll(resp.Body))

	// Extract and print the title from the HTML
	title := extractTitle(string(body))
	fmt.Printf("%s\n", title)

	// Output:
	// Google Public DNS
}

// extractTitle extracts the content of the <title> tag from HTML.
func extractTitle(html string) string {
	const startTag = "<title>"
	const endTag = "</title>"
	start := strings.Index(html, startTag)
	if start == -1 {
		return ""
	}
	start += len(startTag)
	end := strings.Index(html[start:], endTag)
	if end == -1 {
		return ""
	}
	return html[start : start+end]
}
