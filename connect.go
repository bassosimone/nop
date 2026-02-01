//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Adapted from: https://github.com/ooni/probe-cli/blob/v3.20.1/internal/netxlite/dialer.go
// Adapted from: https://github.com/rbmk-project/rbmk/blob/v0.17.0/pkg/x/netcore/dialer.go
//

package nop

import (
	"context"
	"log/slog"
	"net"
	"net/netip"
	"time"

	"github.com/bassosimone/safeconn"
)

// Dialer abstracts the [*net.Dialer] behavior.
//
// By making [*ConnectFunc] depend on an abstract implementation we
// allow for unit testing and for using alternative dialers.
type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// NewConnectFunc returns a new [*ConnectFunc] with default dialer.
//
// The cfg argument contains the common configuration for nop operations.
//
// The network argument must be either "tcp" or "udp".
//
// The logger argument is the [SLogger] to use for structured logging.
func NewConnectFunc(cfg *Config, network string, logger SLogger) *ConnectFunc {
	return &ConnectFunc{
		Dialer:        cfg.Dialer,
		ErrClassifier: cfg.ErrClassifier,
		Logger:        logger,
		Network:       network,
		TimeNow:       cfg.TimeNow,
	}
}

// ConnectFunc dials a [netip.AddrPort] using a configured network.
//
// Returns either a valid [net.Conn] or an error, never both.
//
// All fields are safe to modify after construction but before first use.
// Fields must not be mutated concurrently with calls to [Call].
type ConnectFunc struct {
	// Dialer is the [Dialer] to use.
	//
	// Set by [NewConnectFunc] from [Config.Dialer].
	Dialer Dialer

	// ErrClassifier classifies errors for structured logging.
	//
	// Set by [NewConnectFunc] from [Config.ErrClassifier].
	ErrClassifier ErrClassifier

	// Logger is the [SLogger] to use (configurable for testing or custom logging).
	//
	// Set by [NewConnectFunc] to the user-provided logger.
	Logger SLogger

	// Network is the network to use (either "tcp" or "udp").
	//
	// Set by [NewConnectFunc] to the user-provided value.
	Network string

	// TimeNow is the function to get the current time (configurable for testing).
	//
	// Set by [NewConnectFunc] from [Config.TimeNow].
	TimeNow func() time.Time
}

var _ Func[netip.AddrPort, net.Conn] = &ConnectFunc{}

// Call invokes the [*ConnectFunc] to connect to the given [netip.AddrPort].
func (op *ConnectFunc) Call(ctx context.Context, address netip.AddrPort) (net.Conn, error) {
	t0 := op.TimeNow()
	deadline, _ := ctx.Deadline()
	op.logConnectStart(op.Network, address.String(), t0, deadline)
	conn, err := op.Dialer.DialContext(ctx, op.Network, address.String())
	op.logConnectDone(op.Network, address.String(), t0, deadline, conn, err)
	return conn, err
}

func (op *ConnectFunc) logConnectStart(network, address string, t0 time.Time, deadline time.Time) {
	op.Logger.Info(
		"connectStart",
		slog.Time("deadline", deadline),
		slog.String("protocol", network),
		slog.String("remoteAddr", address),
		slog.Time("t", t0),
	)
}

func (op *ConnectFunc) logConnectDone(
	network, address string, t0 time.Time, deadline time.Time, conn net.Conn, err error) {
	op.Logger.Info(
		"connectDone",
		slog.Time("deadline", deadline),
		slog.Any("err", err),
		slog.String("errClass", op.ErrClassifier.Classify(err)),
		slog.String("localAddr", safeconn.LocalAddr(conn)),
		slog.String("protocol", network),
		slog.String("remoteAddr", address),
		slog.Time("t0", t0),
		slog.Time("t", op.TimeNow()),
	)
}
