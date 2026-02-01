// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompose2(t *testing.T) {
	t.Run("success path", func(t *testing.T) {
		op1 := FuncAdapter[int, string](func(ctx context.Context, n int) (string, error) {
			return "hello", nil
		})
		op2 := FuncAdapter[string, int](func(ctx context.Context, s string) (int, error) {
			return len(s), nil
		})

		composed := Compose2[int, string, int](op1, op2)
		result, err := composed.Call(context.Background(), 42)

		require.NoError(t, err)
		assert.Equal(t, 5, result) // len("hello") = 5
	})

	t.Run("first operation fails", func(t *testing.T) {
		wantErr := errors.New("op1 failed")
		op1 := FuncAdapter[int, string](func(ctx context.Context, n int) (string, error) {
			return "", wantErr
		})
		op2 := FuncAdapter[string, int](func(ctx context.Context, s string) (int, error) {
			t.Fatal("op2 should not be called")
			return 0, nil
		})

		composed := Compose2[int, string, int](op1, op2)
		_, err := composed.Call(context.Background(), 42)

		require.ErrorIs(t, err, wantErr)
	})

	t.Run("second operation fails", func(t *testing.T) {
		wantErr := errors.New("op2 failed")
		op1 := FuncAdapter[int, string](func(ctx context.Context, n int) (string, error) {
			return "hello", nil
		})
		op2 := FuncAdapter[string, int](func(ctx context.Context, s string) (int, error) {
			return 0, wantErr
		})

		composed := Compose2[int, string, int](op1, op2)
		_, err := composed.Call(context.Background(), 42)

		require.ErrorIs(t, err, wantErr)
	})
}

func TestCompose3(t *testing.T) {
	op1 := FuncAdapter[int, int](func(ctx context.Context, n int) (int, error) {
		return n + 1, nil
	})
	op2 := FuncAdapter[int, int](func(ctx context.Context, n int) (int, error) {
		return n * 2, nil
	})
	op3 := FuncAdapter[int, int](func(ctx context.Context, n int) (int, error) {
		return n - 3, nil
	})

	composed := Compose3[int, int, int, int](op1, op2, op3)
	result, err := composed.Call(context.Background(), 5)

	require.NoError(t, err)
	// (5 + 1) * 2 - 3 = 12 - 3 = 9
	assert.Equal(t, 9, result)
}

func TestCompose4(t *testing.T) {
	op1 := FuncAdapter[int, int](func(ctx context.Context, n int) (int, error) { return n + 1, nil })
	op2 := FuncAdapter[int, int](func(ctx context.Context, n int) (int, error) { return n + 1, nil })
	op3 := FuncAdapter[int, int](func(ctx context.Context, n int) (int, error) { return n + 1, nil })
	op4 := FuncAdapter[int, int](func(ctx context.Context, n int) (int, error) { return n + 1, nil })

	composed := Compose4[int, int, int, int, int](op1, op2, op3, op4)
	result, err := composed.Call(context.Background(), 0)

	require.NoError(t, err)
	assert.Equal(t, 4, result)
}

func TestCompose5(t *testing.T) {
	op := FuncAdapter[int, int](func(ctx context.Context, n int) (int, error) { return n + 1, nil })

	composed := Compose5[int, int, int, int, int, int](op, op, op, op, op)
	result, err := composed.Call(context.Background(), 0)

	require.NoError(t, err)
	assert.Equal(t, 5, result)
}

func TestCompose6(t *testing.T) {
	op := FuncAdapter[int, int](func(ctx context.Context, n int) (int, error) { return n + 1, nil })

	composed := Compose6[int, int, int, int, int, int, int](op, op, op, op, op, op)
	result, err := composed.Call(context.Background(), 0)

	require.NoError(t, err)
	assert.Equal(t, 6, result)
}

func TestCompose7(t *testing.T) {
	op := FuncAdapter[int, int](func(ctx context.Context, n int) (int, error) { return n + 1, nil })

	composed := Compose7[int, int, int, int, int, int, int, int](op, op, op, op, op, op, op)
	result, err := composed.Call(context.Background(), 0)

	require.NoError(t, err)
	assert.Equal(t, 7, result)
}

func TestCompose8(t *testing.T) {
	op := FuncAdapter[int, int](func(ctx context.Context, n int) (int, error) { return n + 1, nil })

	composed := Compose8[int, int, int, int, int, int, int, int, int](op, op, op, op, op, op, op, op)
	result, err := composed.Call(context.Background(), 0)

	require.NoError(t, err)
	assert.Equal(t, 8, result)
}

func TestApply(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		fn := FuncAdapter[string, int](func(ctx context.Context, s string) (int, error) {
			return len(s), nil
		})

		applied := Apply(fn, "hello")
		result, err := applied.Call(context.Background(), Unit{})

		require.NoError(t, err)
		assert.Equal(t, 5, result)
	})

	t.Run("error case", func(t *testing.T) {
		wantErr := errors.New("failed")
		fn := FuncAdapter[string, int](func(ctx context.Context, s string) (int, error) {
			return 0, wantErr
		})

		applied := Apply(fn, "hello")
		_, err := applied.Call(context.Background(), Unit{})

		require.ErrorIs(t, err, wantErr)
	})
}

func TestConstFunc(t *testing.T) {
	t.Run("returns constant string", func(t *testing.T) {
		cf := ConstFunc("constant value")
		result, err := cf.Call(context.Background(), Unit{})

		require.NoError(t, err)
		assert.Equal(t, "constant value", result)
	})

	t.Run("returns constant int", func(t *testing.T) {
		cf := ConstFunc(42)
		result, err := cf.Call(context.Background(), Unit{})

		require.NoError(t, err)
		assert.Equal(t, 42, result)
	})

	t.Run("returns constant struct", func(t *testing.T) {
		type myStruct struct {
			X int
			Y string
		}
		want := myStruct{X: 10, Y: "test"}

		cf := ConstFunc(want)
		result, err := cf.Call(context.Background(), Unit{})

		require.NoError(t, err)
		assert.Equal(t, want, result)
	})
}
