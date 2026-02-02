# Composable Network Measurement Primitives

[![GoDoc](https://pkg.go.dev/badge/github.com/bassosimone/nop)](https://pkg.go.dev/github.com/bassosimone/nop) [![Build Status](https://github.com/bassosimone/nop/actions/workflows/go.yml/badge.svg)](https://github.com/bassosimone/nop/actions) [![codecov](https://codecov.io/gh/bassosimone/nop/branch/main/graph/badge.svg)](https://codecov.io/gh/bassosimone/nop)

The `nop` Go package provides composable primitives for building network
measurement pipelines with structured logging. Each primitive is a
`Func[A, B]` that can be chained via type-safe composition (`Compose2`
through `Compose8`).

Basic usage (DNS-over-UDP lookup):

```Go
import (
	"context"
	"log/slog"
	"net/netip"
	"os"
	"time"

	"github.com/bassosimone/nop"
)

// Create configuration with default settings. To classify errors
// in log events, set cfg.ErrClassifier to your own implementation.
cfg := nop.NewConfig()
// cfg.ErrClassifier = nop.ErrClassifierFunc(myClassifyFunc)

// Use LevelDebug to include per-I/O events (read, write, deadline);
// use LevelInfo to see only lifecycle and protocol events.
logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
	Level: slog.LevelDebug,
}))

// Attach a span ID (UUIDv7) so all log entries from this operation
// can be correlated across pipeline stages.
logger = logger.With("spanID", nop.NewSpanID())

pipeline := nop.Compose5(
	nop.NewEndpointFunc(netip.MustParseAddrPort("8.8.8.8:53")),
	nop.NewConnectFunc(cfg, "udp", logger),
	nop.NewObserveConnFunc(cfg, logger),
	nop.NewCancelWatchFunc(),
	nop.NewDNSOverUDPConnFunc(cfg, logger),
)

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

dnsConn, err := pipeline.Call(ctx, nop.Unit{})
// ... use dnsConn.Exchange(ctx, query)
```

See the [package documentation](https://pkg.go.dev/github.com/bassosimone/nop)
and testable examples for DNS-over-TCP, DNS-over-TLS, DNS-over-HTTPS, and
HTTPS round trip pipelines.

## Installation

To add this package as a dependency to your module:

```sh
go get github.com/bassosimone/nop
```

## Development

To run the tests:

```sh
go test -v .
```

To measure test coverage:

```sh
go test -v -cover .
```

## License

```
SPDX-License-Identifier: GPL-3.0-or-later
```

## History

Adapted from [ooni/probe-cli](https://github.com/ooni/probe-cli/tree/v3.20.1)
and [rbmk-project/rbmk](https://github.com/rbmk-project/rbmk/tree/v0.17.0).
