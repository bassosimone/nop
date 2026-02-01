// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnit(t *testing.T) {
	// Test that Unit zero value is usable
	var u Unit
	assert.Equal(t, Unit{}, u)

	// Test that Unit values are equal
	u1 := Unit{}
	u2 := Unit{}
	assert.Equal(t, u1, u2)
}
