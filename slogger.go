//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Adapted from: https://github.com/ooni/probe-cli/blob/v3.20.1/internal/netxlite/dialer.go
// Adapted from: https://github.com/rbmk-project/rbmk/blob/v0.17.0/pkg/x/netcore/dialer.go
//

package nop

// SLogger abstracts the [*slog.Logger] behavior.
//
// By using an abstraction we allow for unit testing and alternative implementations.
//
// This package uses two log levels:
//   - Info for lifecycle and protocol events (connect, close, TLS handshake,
//     HTTP round trip, DNS exchange, DNS query/response)
//   - Debug for per-I/O events (read, write, set deadline)
//
// The [*slog.Logger] type satisfies this interface.
type SLogger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
}

// DefaultSLogger returns the default [SLogger] to use.
//
// The default is a no-op logger that discards all output. This follows the
// library convention of not writing to stdout/stderr unless explicitly configured.
//
// Use a custom [*slog.Logger] for emitting logs.
func DefaultSLogger() SLogger {
	return discardSLogger{}
}

// discardSLogger is a no-op [SLogger] that discards all log messages.
type discardSLogger struct{}

var _ SLogger = discardSLogger{}

// Debug implements [SLogger].
func (discardSLogger) Debug(msg string, args ...any) {
	// nothing
}

// Info implements [SLogger].
func (discardSLogger) Info(msg string, args ...any) {
	// nothing
}
