// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFuncAdapter(t *testing.T) {
	called := false
	adapter := FuncAdapter[int, string](func(ctx context.Context, input int) (string, error) {
		called = true
		return "result", nil
	})

	output, err := adapter.Call(context.Background(), 42)

	require.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "result", output)
}
