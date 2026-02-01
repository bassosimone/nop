//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Adapted from: https://github.com/ooni/probe-cli/blob/v3.20.0/internal/x/dslx/fxasync.go
// Adapted from: https://github.com/ooni/probe-cli/blob/v3.20.0/internal/x/dslx/fxcore.go
// Adapted from: https://github.com/ooni/probe-cli/blob/v3.20.0/internal/x/dslx/fxstream.go
//

package nop

import "context"

// Compose2 chains two [Func] instances together into a pipeline.
//
// The output of op1 becomes the input to op2. If op1 returns an error,
// op2 is not called and the error is returned immediately.
func Compose2[A, B, C any](op1 Func[A, B], op2 Func[B, C]) Func[A, C] {
	return &compose2[A, B, C]{op1, op2}
}

type compose2[A, B, C any] struct {
	op1 Func[A, B]
	op2 Func[B, C]
}

func (c *compose2[A, B, C]) Call(ctx context.Context, input A) (C, error) {
	res, err := c.op1.Call(ctx, input)
	if err != nil {
		var zero C
		return zero, err
	}
	return c.op2.Call(ctx, res)
}

// Compose3 chains three [Func] instances together.
func Compose3[A, B, C, D any](op1 Func[A, B], op2 Func[B, C], op3 Func[C, D]) Func[A, D] {
	return Compose2(op1, Compose2(op2, op3))
}

// Compose4 chains four [Func] instances together.
func Compose4[A, B, C, D, E any](op1 Func[A, B], op2 Func[B, C], op3 Func[C, D], op4 Func[D, E]) Func[A, E] {
	return Compose2(op1, Compose3(op2, op3, op4))
}

// Compose5 chains five [Func] instances together.
func Compose5[A, B, C, D, E, F any](op1 Func[A, B], op2 Func[B, C], op3 Func[C, D], op4 Func[D, E], op5 Func[E, F]) Func[A, F] {
	return Compose2(op1, Compose4(op2, op3, op4, op5))
}

// Compose6 chains six [Func] instances together.
func Compose6[A, B, C, D, E, F, G any](
	op1 Func[A, B], op2 Func[B, C], op3 Func[C, D], op4 Func[D, E], op5 Func[E, F], op6 Func[F, G]) Func[A, G] {
	return Compose2(op1, Compose5(op2, op3, op4, op5, op6))
}

// Compose7 chains seven [Func] instances together.
func Compose7[A, B, C, D, E, F, G, H any](
	op1 Func[A, B], op2 Func[B, C], op3 Func[C, D], op4 Func[D, E], op5 Func[E, F], op6 Func[F, G], op7 Func[G, H]) Func[A, H] {
	return Compose2(op1, Compose6(op2, op3, op4, op5, op6, op7))
}

// Compose8 chains eight [Func] instances together.
func Compose8[A, B, C, D, E, F, G, H, I any](op1 Func[A, B],
	op2 Func[B, C], op3 Func[C, D], op4 Func[D, E], op5 Func[E, F], op6 Func[F, G], op7 Func[G, H], op8 Func[H, I]) Func[A, I] {
	return Compose2(op1, Compose7(op2, op3, op4, op5, op6, op7, op8))
}

// Apply binds a fixed input to a [Func], returning a [Func] that takes [Unit] instead.
//
// This is useful for currying a pipeline that requires an input value into a
// pipeline that can be used where a [Func[Unit, B]] is expected.
func Apply[A, B any](fn Func[A, B], input A) Func[Unit, B] {
	return &apply[A, B]{fn, input}
}

type apply[A, B any] struct {
	fn    Func[A, B]
	input A
}

func (b *apply[A, B]) Call(ctx context.Context, _ Unit) (B, error) {
	return b.fn.Call(ctx, b.input)
}

// ConstFunc returns a [Func] that always returns the given value.
//
// This lifts a pure value into the [Func] world, creating a [Func[Unit, B]]
// that ignores its input and returns the constant value.
func ConstFunc[B any](value B) Func[Unit, B] {
	return &constFunc[B]{value}
}

type constFunc[B any] struct {
	value B
}

func (c *constFunc[B]) Call(ctx context.Context, _ Unit) (B, error) {
	return c.value, nil
}
