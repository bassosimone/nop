// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultErrClassifier(t *testing.T) {
	// Should return empty string for nil
	result := DefaultErrClassifier.Classify(nil)
	assert.Equal(t, "", result)

	// Should return empty string for any error
	result = DefaultErrClassifier.Classify(errors.New("some error"))
	assert.Equal(t, "", result)
}
