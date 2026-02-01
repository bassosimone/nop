// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	require.NotNil(t, cfg)

	// Dialer should be set to *net.Dialer
	_, ok := cfg.Dialer.(*net.Dialer)
	assert.True(t, ok, "Dialer should be *net.Dialer")

	// ErrClassifier should be DefaultErrClassifier
	assert.Equal(t, "", cfg.ErrClassifier.Classify(nil))

	// TimeNow should be set and return a valid time
	now := cfg.TimeNow()
	assert.False(t, now.IsZero())
}
