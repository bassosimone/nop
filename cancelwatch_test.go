// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"context"
	"testing"
	"time"

	"github.com/bassosimone/netstub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NewCancelWatchFunc returns a non-nil value.
func TestNewCancelWatchFunc(t *testing.T) {
	fn := NewCancelWatchFunc()
	require.NotNil(t, fn)
}

// Call returns a wrapped conn that delegates Close to the underlying conn.
func TestCancelWatchFuncCall(t *testing.T) {
	fn := NewCancelWatchFunc()

	closeCalled := false
	mockConn := &netstub.FuncConn{
		CloseFunc: func() error {
			closeCalled = true
			return nil
		},
	}

	result, err := fn.Call(context.Background(), mockConn)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Closing the wrapper delegates to the underlying conn.
	err = result.Close()
	require.NoError(t, err)
	assert.True(t, closeCalled)
}

// Cancelling the context triggers Close on the underlying conn.
func TestCancelWatchFuncClosesOnCancel(t *testing.T) {
	fn := NewCancelWatchFunc()

	done := make(chan bool, 1)
	mockConn := &netstub.FuncConn{
		CloseFunc: func() error {
			done <- true
			return nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	_, err := fn.Call(ctx, mockConn)
	require.NoError(t, err)

	// Connection not closed before cancelling the context.
	select {
	case <-done:
		t.Fatal("connection should not be closed yet")
	default:
	}

	cancel()

	// Wait for AfterFunc to close the connection.
	waitClose := func() bool {
		return <-done
	}
	assert.Eventually(t, waitClose, 1*time.Second, 10*time.Millisecond)
}

// If the context is already cancelled, the connection is closed immediately.
func TestCancelWatchFuncAlreadyCancelled(t *testing.T) {
	fn := NewCancelWatchFunc()

	done := make(chan bool, 1)
	mockConn := &netstub.FuncConn{
		CloseFunc: func() error {
			done <- true
			return nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := fn.Call(ctx, mockConn)
	require.NoError(t, err)

	// Wait for AfterFunc to see the already-cancelled context and close.
	waitClose := func() bool {
		return <-done
	}
	assert.Eventually(t, waitClose, 1*time.Second, 10*time.Millisecond)
}

// Closing the wrapper unregisters the watcher so that subsequent context
// cancellation does not call Close on the underlying conn a second time.
func TestCancelWatchFuncCloseUnregistersWatcher(t *testing.T) {
	fn := NewCancelWatchFunc()

	closeCount := 0
	mockConn := &netstub.FuncConn{
		CloseFunc: func() error {
			closeCount++
			return nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result, err := fn.Call(ctx, mockConn)
	require.NoError(t, err)

	// Close the wrapper — should unregister the watcher and close the conn.
	err = result.Close()
	require.NoError(t, err)
	assert.Equal(t, 1, closeCount)

	// Cancel the context — should NOT trigger another close.
	cancel()
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, closeCount)
}
