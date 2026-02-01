// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"context"
	"net"
)

// dnsUnusedDialer is a [Dialer] that panics if DialContext is called.
//
// DNS exchange methods use pre-established connections and never dial.
// This type serves as a sentinel to catch programming errors where the
// transport attempts to dial instead of using the provided connection.
type dnsUnusedDialer struct{}

var _ Dialer = dnsUnusedDialer{}

// DialContext implements [Dialer] and always panics.
func (dnsUnusedDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	panic("nop: DNS transport must not dial; this is a programming error")
}
