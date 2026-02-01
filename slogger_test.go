// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultSLogger(t *testing.T) {
	logger := DefaultSLogger()

	// Should return a non-nil logger
	assert.NotNil(t, logger)

	// Should be able to call Debug and Info without panic (discards output)
	logger.Debug("debug message", "key", "value")
	logger.Info("info message", "key", "value")
}

func TestDiscardSLogger(t *testing.T) {
	logger := discardSLogger{}

	// Verify it implements SLogger
	var _ SLogger = logger

	// Should be able to call Debug and Info without panic (discards output)
	logger.Debug("debug message", "key1", "value1", "key2", 42)
	logger.Info("info message", "key1", "value1", "key2", 42)
}
