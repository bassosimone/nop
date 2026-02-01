// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"context"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEndpointFunc(t *testing.T) {
	endpoint := netip.MustParseAddrPort("93.184.216.34:443")

	fn := NewEndpointFunc(endpoint)
	result, err := fn.Call(context.Background(), Unit{})

	require.NoError(t, err)
	assert.Equal(t, endpoint, result)
}

func TestNewEndpointFuncIPv6(t *testing.T) {
	endpoint := netip.MustParseAddrPort("[2001:db8::1]:8080")

	fn := NewEndpointFunc(endpoint)
	result, err := fn.Call(context.Background(), Unit{})

	require.NoError(t, err)
	assert.Equal(t, endpoint, result)
}
