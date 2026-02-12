//go:build windows

//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Adapted from: https://github.com/rbmk-project/rbmk/blob/v0.17.0/pkg/common/errclass/windows.go
//

package errclass

import "golang.org/x/sys/windows"

const (
	errEADDRNOTAVAIL   = windows.WSAEADDRNOTAVAIL
	errEADDRINUSE      = windows.WSAEADDRINUSE
	errECONNABORTED    = windows.WSAECONNABORTED
	errECONNREFUSED    = windows.WSAECONNREFUSED
	errECONNRESET      = windows.WSAECONNRESET
	errEHOSTUNREACH    = windows.WSAEHOSTUNREACH
	errEINVAL          = windows.WSAEINVAL
	errEINTR           = windows.WSAEINTR
	errENETDOWN        = windows.WSAENETDOWN
	errENETUNREACH     = windows.WSAENETUNREACH
	errENOBUFS         = windows.WSAENOBUFS
	errENOTCONN        = windows.WSAENOTCONN
	errEPROTONOSUPPORT = windows.WSAEPROTONOSUPPORT
	errETIMEDOUT       = windows.WSAETIMEDOUT
)
