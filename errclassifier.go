// SPDX-License-Identifier: GPL-3.0-or-later

package nop

// ErrClassifier classifies errors into categorical strings for analysis.
//
// Implementations map errors to short, descriptive labels (e.g., "ETIMEDOUT",
// "ECONNRESET") that facilitate systematic analysis of network measurement results.
type ErrClassifier interface {
	Classify(err error) string
}

// ErrClassifierFunc adapts a function to the [ErrClassifier] interface.
//
// This allows using simple functions as classifiers:
//
//	op.ErrClassifier = ErrClassifierFunc(errclass.New)
type ErrClassifierFunc func(error) string

var _ ErrClassifier = ErrClassifierFunc(nil)

// Classify implements [ErrClassifier].
func (f ErrClassifierFunc) Classify(err error) string {
	return f(err)
}

// DefaultErrClassifier is a no-op classifier that returns an empty string.
var DefaultErrClassifier = ErrClassifierFunc(func(error) string { return "" })
