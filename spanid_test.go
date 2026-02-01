// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSpanID(t *testing.T) {
	spanID := NewSpanID()

	// Should be a valid UUID string
	parsed, err := uuid.Parse(spanID)
	require.NoError(t, err)

	// Should be version 7 (time-ordered)
	assert.Equal(t, uuid.Version(7), parsed.Version())
}

func TestNewSpanIDUniqueness(t *testing.T) {
	// Generate multiple span IDs and verify they're all unique
	const count = 100
	seen := make(map[string]struct{}, count)

	for range count {
		spanID := NewSpanID()
		_, duplicate := seen[spanID]
		require.False(t, duplicate, "duplicate span ID generated: %s", spanID)
		seen[spanID] = struct{}{}
	}
}
