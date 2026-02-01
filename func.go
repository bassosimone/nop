// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import "context"

// Func is a generic operation that accepts an input and returns a result.
//
// Func instances can be composed using [Compose2], [Compose3], etc. to create
// type-safe pipelines where the output of one operation flows to the input of the next.
//
// Resource cleanup contract: when a Func receives a closeable resource as input
// and returns an error, it is responsible for closing that resource before returning.
// This ensures that composed pipelines do not leak resources on partial failure.
// See [TLSHandshakeFunc] for an example of this pattern.
type Func[A, B any] interface {
	Call(ctx context.Context, input A) (B, error)
}

// FuncAdapter wraps a function as a [Func] implementation.
//
// Use this to create ad-hoc [Func] instances from closures when you need
// custom behavior that doesn't fit the existing primitives.
type FuncAdapter[A, B any] func(ctx context.Context, input A) (B, error)

// Call implements [Func].
func (f FuncAdapter[A, B]) Call(ctx context.Context, input A) (B, error) {
	return f(ctx, input)
}
