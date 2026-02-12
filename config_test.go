// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"context"
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

	// ErrClassifier should use errclass by default
	assert.Equal(t, "", cfg.ErrClassifier.Classify(nil))
	assert.Equal(t, "ETIMEDOUT", cfg.ErrClassifier.Classify(context.DeadlineExceeded))

	// TimeNow should be set and return a valid time
	now := cfg.TimeNow()
	assert.False(t, now.IsZero())
}
