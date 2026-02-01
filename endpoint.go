// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import "net/netip"

// NewEndpointFunc returns a [Func] that always returns the given [netip.AddrPort].
//
// This is a convenience wrapper around [ConstFunc] for the common case of
// injecting a network endpoint into a pipeline.
func NewEndpointFunc(endpoint netip.AddrPort) Func[Unit, netip.AddrPort] {
	return ConstFunc(endpoint)
}
