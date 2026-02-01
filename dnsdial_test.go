// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// dnsUnusedDialer panics when DialContext is called.
func TestDNSUnusedDialerPanics(t *testing.T) {
	d := dnsUnusedDialer{}
	assert.Panics(t, func() {
		d.DialContext(context.Background(), "tcp", "127.0.0.1:53")
	})
}
