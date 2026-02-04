// SPDX-License-Identifier: GPL-3.0-or-later

// Package nop provides composable primitives for network measurement pipelines.
//
// # Core Abstraction
//
// The package is built around a single interface:
//
//	type Func[A, B any] interface {
//		Call(ctx context.Context, input A) (B, error)
//	}
//
// Each Func represents an atomic network operation with exactly one success
// mode and one failure mode. This design enables type-safe composition via
// [Compose2], [Compose3], etc., where the compiler verifies that outputs
// match inputs across pipeline stages.
//
// # Available Primitives
//
// Connection establishment:
//   - [ConnectFunc]: dials TCP or UDP endpoints
//   - [TLSHandshakeFunc]: performs TLS handshake over an existing connection
//   - [ObserveConnFunc]: observes connections for logging I/O operations
//   - [CancelWatchFunc]: closes connection on context cancellation (for responsive ^C handling)
//
// HTTP:
//   - [HTTPConn]: wraps a connection with an HTTP transport, performs round trips
//     with structured logging and transparent body observation (created via [NewHTTPConnFunc])
//
// DNS resolution:
//   - [DNSOverUDPConn]: wraps a UDP connection for DNS-over-UDP (owns the connection)
//   - [DNSOverTCPConn]: wraps a TCP connection for DNS-over-TCP (owns the connection)
//   - [DNSOverTLSConn]: wraps a TLS connection for DNS-over-TLS (owns the connection)
//   - [DNSOverHTTPSConn]: wraps an HTTPConn for DNS-over-HTTPS (owns the connection)
//   - [DNSExchangeLogContext]: structured logging for DNS exchanges, used internally
//     by the above types and available for callers implementing custom exchange
//     loops (e.g., collecting duplicate DNS-over-UDP responses)
//
// Composition utilities:
//   - [Compose2] through [Compose8]: chain Funcs into pipelines
//   - [FuncAdapter]: wrap a function as a Func for ad-hoc custom behavior
//   - [Apply]: bind a fixed input to a Func
//   - [ConstFunc]: lift a pure value into a Func
//   - [NewEndpointFunc]: convenience wrapper for ConstFunc with endpoints
//
// # Connection Lifecycle
//
// This package uses two ownership patterns for connection management:
//
// Dial operations ([ConnectFunc], [TLSHandshakeFunc]) create connections and
// transfer ownership to the next stage on success. On error, they close the connection.
//
// Wrapper types ([HTTPConn], [DNSOverTLSConn], etc.) OWN their underlying connection.
// The caller must call Close() when done, which closes the underlying connection.
// These can be composed into pipelines via their corresponding Func types.
//
// See the testable examples for complete code demonstrating these patterns.
//
// # Observability
//
// All primitives support structured logging via [SLogger] (compatible with [log/slog]).
//
// By default, logging is disabled. Set the Logger field to a custom [*slog.Logger]
// to enable logging. Error classification is configurable via [ErrClassifier]; by
// default, a no-op classifier is used.
//
// Primitives emit two kinds of structured log events:
//
//   - Span events (*Start/*Done pairs): Record operation lifecycle including
//     timing and success/failure. Used for latency analysis and error tracking.
//
//   - Wire observations (e.g., dnsQuery/dnsResponse): Capture protocol-level
//     messages for dig-like UI output and protocol debugging.
//
// The [SLogger] interface accepts any slog-compatible handler, enabling flexible
// post-processing. Handlers can filter, transform, or route events as needed.
//
// All events share a common set of fields: localAddr, remoteAddr, protocol,
// and t (timestamp). Completion events (*Done) additionally include t0 (start
// time), err, and errClass. I/O-level events (read, write, deadline changes)
// are emitted at [slog.LevelDebug]; all other events use [slog.LevelInfo].
// The structured log format is compatible with the RBMK data format specification
// (see https://github.com/rbmk-project/rbmk) and may evolve in minor ways as
// these packages mature.
//
// Use [NewSpanID] to generate a unique, time-ordered identifier (UUIDv7) for each
// operation, then attach it to the logger with [*slog.Logger.With]. All log entries
// from that operation will share the same spanID, enabling correlation across
// pipeline stages and simplifying log analysis.
//
// # Timeout and Context Philosophy
//
// This package is context-transparent: operations never modify the context they receive.
// The caller controls timeouts externally via [context.WithTimeout], [context.WithDeadline],
// or [signal.NotifyContext]. When the context is done (timeout, cancel, or signal),
// operations fail and the pipeline is interrupted.
//
// Connection lifecycle requires [CancelWatchFunc] to bind the context lifecycle to
// the connection: when the context is done, the connection is closed immediately,
// causing any in-progress I/O to fail. This enables responsive ^C handling via
// [signal.NotifyContext] and ensures that blocking I/O respects the context deadline.
//
// IMPORTANT: Without [CancelWatchFunc] in your pipeline, I/O operations may block
// indefinitely even after the context is done. Always include [CancelWatchFunc]
// when composing connection pipelines to ensure proper timeout behavior.
//
// # Design Boundaries
//
// This package intentionally provides only primitives. The following are out of scope
// and should be implemented by higher-level packages:
//
//   - Parallel execution (fan-out, racing)
//   - Retry and backoff logic
//   - Multi-step orchestration
//   - Convenience helpers that combine multiple primitives
//
// These concerns introduce multiple success/failure modes, which would compromise
// the compositional simplicity of the primitives.
package nop
