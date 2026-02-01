// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"net"
	"time"
)

// Config holds common configuration for nop operations.
//
// Pass this to constructor functions to pre-wire dependencies.
// All fields have sensible defaults set by [NewConfig].
type Config struct {
	// Dialer is used by [*ConnectFunc].
	//
	// Set by [NewConfig] to [*net.Dialer].
	Dialer Dialer

	// ErrClassifier classifies errors for structured logging.
	//
	// Set by [NewConfig] to [DefaultErrClassifier].
	ErrClassifier ErrClassifier

	// TimeNow returns the current time.
	//
	// Set by [NewConfig] to [time.Now].
	TimeNow func() time.Time
}

// NewConfig creates a [*Config] with sensible defaults.
func NewConfig() *Config {
	return &Config{
		Dialer:        &net.Dialer{},
		ErrClassifier: DefaultErrClassifier,
		TimeNow:       time.Now,
	}
}
