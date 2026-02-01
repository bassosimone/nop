// SPDX-License-Identifier: GPL-3.0-or-later

package nop

// Unit is a type not containing any value (analogous to an
// explicit `void` type in C and C++).
//
// Use this type to construct [Func] that take no argument
// or return no value to the caller.
type Unit struct{}
