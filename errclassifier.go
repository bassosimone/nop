// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import "github.com/bassosimone/errclass"

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
//	cfg.ErrClassifier = nop.ErrClassifierFunc(myClassifier)
type ErrClassifierFunc func(error) string

var _ ErrClassifier = ErrClassifierFunc(nil)

// Classify implements [ErrClassifier].
func (f ErrClassifierFunc) Classify(err error) string {
	return f(err)
}

// DefaultErrClassifier uses [errclass.New] to classify errors into
// Unix-like error names (e.g., "ETIMEDOUT", "ECONNRESET", "EDNS_NONAME").
//
// See the [errclass] package for the full list of supported error classes.
var DefaultErrClassifier = ErrClassifierFunc(errclass.New)
