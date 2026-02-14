// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"context"
	"errors"
	"testing"

	"github.com/bassosimone/errclass"
	"github.com/stretchr/testify/assert"
)

func TestDefaultErrClassifier(t *testing.T) {
	// Should return empty string for nil error
	result := DefaultErrClassifier.Classify(nil)
	assert.Equal(t, "", result)

	// Should classify known errors using errclass
	result = DefaultErrClassifier.Classify(context.DeadlineExceeded)
	assert.Equal(t, errclass.ETIMEDOUT, result)

	// Should return EGENERIC for unknown errors
	result = DefaultErrClassifier.Classify(errors.New("unknown error"))
	assert.Equal(t, errclass.EGENERIC, result)
}
