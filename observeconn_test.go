// SPDX-License-Identifier: GPL-3.0-or-later

package nop

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NewObserveConnFunc populates all fields from Config and the provided logger.
func TestNewObserveConnFunc(t *testing.T) {
	cfg := NewConfig()
	logger := DefaultSLogger()

	fn := NewObserveConnFunc(cfg, logger)

	require.NotNil(t, fn)
	assert.NotNil(t, fn.Logger)
	assert.NotNil(t, fn.TimeNow)
	assert.NotNil(t, fn.ErrClassifier)
}

// Call wraps the connection and returns a net.Conn implementation.
func TestObserveConnFunc(t *testing.T) {
	cfg := NewConfig()

	mockConn := newMinimalConn()

	fn := NewObserveConnFunc(cfg, DefaultSLogger())
	observed, err := fn.Call(context.Background(), mockConn)

	require.NoError(t, err)
	require.NotNil(t, observed)

	// Verify it implements net.Conn
	var _ net.Conn = observed
}

// Read delegates to the underlying connection and returns the data.
func TestObservedConnRead(t *testing.T) {
	cfg := NewConfig()

	readData := []byte("hello world")
	mockConn := newMinimalConn()
	mockConn.ReadFunc = func(b []byte) (int, error) {
		copy(b, readData)
		return len(readData), nil
	}

	fn := NewObserveConnFunc(cfg, DefaultSLogger())
	observed, err := fn.Call(context.Background(), mockConn)
	require.NoError(t, err)

	buf := make([]byte, 100)
	n, err := observed.Read(buf)

	require.NoError(t, err)
	assert.Equal(t, len(readData), n)
	assert.Equal(t, readData, buf[:n])
}

// Read propagates errors from the underlying connection.
func TestObservedConnReadError(t *testing.T) {
	cfg := NewConfig()
	wantErr := errors.New("read error")

	mockConn := newMinimalConn()
	mockConn.ReadFunc = func(b []byte) (int, error) {
		return 0, wantErr
	}

	fn := NewObserveConnFunc(cfg, DefaultSLogger())
	observed, _ := fn.Call(context.Background(), mockConn)

	buf := make([]byte, 100)
	_, err := observed.Read(buf)

	require.ErrorIs(t, err, wantErr)
}

// Write delegates to the underlying connection and sends the data.
func TestObservedConnWrite(t *testing.T) {
	cfg := NewConfig()

	var writtenData []byte
	mockConn := newMinimalConn()
	mockConn.WriteFunc = func(b []byte) (int, error) {
		writtenData = append(writtenData, b...)
		return len(b), nil
	}

	fn := NewObserveConnFunc(cfg, DefaultSLogger())
	observed, err := fn.Call(context.Background(), mockConn)
	require.NoError(t, err)

	data := []byte("test data")
	n, err := observed.Write(data)

	require.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, data, writtenData)
}

// Write propagates errors from the underlying connection.
func TestObservedConnWriteError(t *testing.T) {
	cfg := NewConfig()
	wantErr := errors.New("write error")

	mockConn := newMinimalConn()
	mockConn.WriteFunc = func(b []byte) (int, error) {
		return 0, wantErr
	}

	fn := NewObserveConnFunc(cfg, DefaultSLogger())
	observed, _ := fn.Call(context.Background(), mockConn)

	_, err := observed.Write([]byte("test"))

	require.ErrorIs(t, err, wantErr)
}

// Second Close returns net.ErrClosed without calling the underlying Close again.
func TestObservedConnCloseOnce(t *testing.T) {
	cfg := NewConfig()

	closeCount := 0
	mockConn := newMinimalConn()
	mockConn.CloseFunc = func() error {
		closeCount++
		return nil
	}

	fn := NewObserveConnFunc(cfg, DefaultSLogger())
	observed, _ := fn.Call(context.Background(), mockConn)

	// First close should work
	err1 := observed.Close()
	require.NoError(t, err1)
	assert.Equal(t, 1, closeCount)

	// Second close should return ErrClosed without calling underlying Close
	err2 := observed.Close()
	require.ErrorIs(t, err2, net.ErrClosed)
	assert.Equal(t, 1, closeCount) // Still 1
}

// LocalAddr delegates to the underlying connection.
func TestObservedConnLocalAddr(t *testing.T) {
	cfg := NewConfig()
	wantAddr := &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 54321}

	mockConn := newMinimalConn()
	mockConn.LocalAddrFunc = func() net.Addr { return wantAddr }

	fn := NewObserveConnFunc(cfg, DefaultSLogger())
	observed, _ := fn.Call(context.Background(), mockConn)

	assert.Equal(t, wantAddr, observed.LocalAddr())
}

// RemoteAddr delegates to the underlying connection.
func TestObservedConnRemoteAddr(t *testing.T) {
	cfg := NewConfig()
	wantAddr := &net.TCPAddr{IP: net.IPv4(93, 184, 216, 34), Port: 443}

	mockConn := newMinimalConn()
	mockConn.RemoteAddrFunc = func() net.Addr { return wantAddr }

	fn := NewObserveConnFunc(cfg, DefaultSLogger())
	observed, _ := fn.Call(context.Background(), mockConn)

	assert.Equal(t, wantAddr, observed.RemoteAddr())
}

// SetDeadline delegates to the underlying connection.
func TestObservedConnSetDeadline(t *testing.T) {
	cfg := NewConfig()
	wantDeadline := time.Now().Add(time.Hour)
	var gotDeadline time.Time

	mockConn := newMinimalConn()
	mockConn.SetDeadlineFunc = func(t time.Time) error {
		gotDeadline = t
		return nil
	}

	fn := NewObserveConnFunc(cfg, DefaultSLogger())
	observed, _ := fn.Call(context.Background(), mockConn)

	err := observed.SetDeadline(wantDeadline)

	require.NoError(t, err)
	assert.Equal(t, wantDeadline, gotDeadline)
}

// SetReadDeadline delegates to the underlying connection.
func TestObservedConnSetReadDeadline(t *testing.T) {
	cfg := NewConfig()
	wantDeadline := time.Now().Add(time.Hour)
	var gotDeadline time.Time

	mockConn := newMinimalConn()
	mockConn.SetReadDeadFunc = func(t time.Time) error {
		gotDeadline = t
		return nil
	}

	fn := NewObserveConnFunc(cfg, DefaultSLogger())
	observed, _ := fn.Call(context.Background(), mockConn)

	err := observed.SetReadDeadline(wantDeadline)

	require.NoError(t, err)
	assert.Equal(t, wantDeadline, gotDeadline)
}

// SetWriteDeadline delegates to the underlying connection.
func TestObservedConnSetWriteDeadline(t *testing.T) {
	cfg := NewConfig()
	wantDeadline := time.Now().Add(time.Hour)
	var gotDeadline time.Time

	mockConn := newMinimalConn()
	mockConn.SetWriteDeaFunc = func(t time.Time) error {
		gotDeadline = t
		return nil
	}

	fn := NewObserveConnFunc(cfg, DefaultSLogger())
	observed, _ := fn.Call(context.Background(), mockConn)

	err := observed.SetWriteDeadline(wantDeadline)

	require.NoError(t, err)
	assert.Equal(t, wantDeadline, gotDeadline)
}

// Close emits closeStart/closeDone log events.
func TestObservedConnCloseLogging(t *testing.T) {
	cfg := NewConfig()
	logger, records := newCapturingLogger()

	mockConn := newMinimalConn()
	mockConn.CloseFunc = func() error { return nil }

	fn := NewObserveConnFunc(cfg, logger)
	observed, _ := fn.Call(context.Background(), mockConn)

	_ = observed.Close()

	require.Len(t, *records, 2)
	assert.Equal(t, "closeStart", (*records)[0].Message)
	assert.Equal(t, "closeDone", (*records)[1].Message)
}

// Read emits readStart/readDone log events.
func TestObservedConnReadLogging(t *testing.T) {
	cfg := NewConfig()
	logger, records := newCapturingLogger()

	mockConn := newMinimalConn()
	mockConn.ReadFunc = func(b []byte) (int, error) { return 0, nil }

	fn := NewObserveConnFunc(cfg, logger)
	observed, _ := fn.Call(context.Background(), mockConn)

	buf := make([]byte, 10)
	_, _ = observed.Read(buf)

	require.Len(t, *records, 2)
	assert.Equal(t, "readStart", (*records)[0].Message)
	assert.Equal(t, "readDone", (*records)[1].Message)
}

// Write emits writeStart/writeDone log events.
func TestObservedConnWriteLogging(t *testing.T) {
	cfg := NewConfig()
	logger, records := newCapturingLogger()

	mockConn := newMinimalConn()
	mockConn.WriteFunc = func(b []byte) (int, error) { return len(b), nil }

	fn := NewObserveConnFunc(cfg, logger)
	observed, _ := fn.Call(context.Background(), mockConn)

	_, _ = observed.Write([]byte("test"))

	require.Len(t, *records, 2)
	assert.Equal(t, "writeStart", (*records)[0].Message)
	assert.Equal(t, "writeDone", (*records)[1].Message)
}

// SetDeadline propagates errors from the underlying connection.
func TestObservedConnSetDeadlineError(t *testing.T) {
	cfg := NewConfig()
	wantErr := errors.New("set deadline error")

	mockConn := newMinimalConn()
	mockConn.SetDeadlineFunc = func(time.Time) error {
		return wantErr
	}

	fn := NewObserveConnFunc(cfg, DefaultSLogger())
	observed, _ := fn.Call(context.Background(), mockConn)

	err := observed.SetDeadline(time.Now().Add(time.Hour))

	require.ErrorIs(t, err, wantErr)
}

// SetReadDeadline propagates errors from the underlying connection.
func TestObservedConnSetReadDeadlineError(t *testing.T) {
	cfg := NewConfig()
	wantErr := errors.New("set read deadline error")

	mockConn := newMinimalConn()
	mockConn.SetReadDeadFunc = func(time.Time) error {
		return wantErr
	}

	fn := NewObserveConnFunc(cfg, DefaultSLogger())
	observed, _ := fn.Call(context.Background(), mockConn)

	err := observed.SetReadDeadline(time.Now().Add(time.Hour))

	require.ErrorIs(t, err, wantErr)
}

// SetWriteDeadline propagates errors from the underlying connection.
func TestObservedConnSetWriteDeadlineError(t *testing.T) {
	cfg := NewConfig()
	wantErr := errors.New("set write deadline error")

	mockConn := newMinimalConn()
	mockConn.SetWriteDeaFunc = func(time.Time) error {
		return wantErr
	}

	fn := NewObserveConnFunc(cfg, DefaultSLogger())
	observed, _ := fn.Call(context.Background(), mockConn)

	err := observed.SetWriteDeadline(time.Now().Add(time.Hour))

	require.ErrorIs(t, err, wantErr)
}

// Close propagates errors from the underlying connection on the first call.
func TestObservedConnCloseError(t *testing.T) {
	cfg := NewConfig()
	wantErr := errors.New("close error")

	mockConn := newMinimalConn()
	mockConn.CloseFunc = func() error {
		return wantErr
	}

	fn := NewObserveConnFunc(cfg, DefaultSLogger())
	observed, _ := fn.Call(context.Background(), mockConn)

	err := observed.Close()

	require.ErrorIs(t, err, wantErr)
}

// SetDeadline emits a setDeadline log event.
func TestObservedConnSetDeadlineLogging(t *testing.T) {
	cfg := NewConfig()
	logger, records := newCapturingLogger()

	mockConn := newMinimalConn()
	mockConn.SetDeadlineFunc = func(time.Time) error { return nil }

	fn := NewObserveConnFunc(cfg, logger)
	observed, _ := fn.Call(context.Background(), mockConn)

	_ = observed.SetDeadline(time.Now().Add(time.Hour))

	require.Len(t, *records, 1)
	assert.Equal(t, "setDeadline", (*records)[0].Message)
}

// SetReadDeadline emits a setReadDeadline log event.
func TestObservedConnSetReadDeadlineLogging(t *testing.T) {
	cfg := NewConfig()
	logger, records := newCapturingLogger()

	mockConn := newMinimalConn()
	mockConn.SetReadDeadFunc = func(time.Time) error { return nil }

	fn := NewObserveConnFunc(cfg, logger)
	observed, _ := fn.Call(context.Background(), mockConn)

	_ = observed.SetReadDeadline(time.Now().Add(time.Hour))

	require.Len(t, *records, 1)
	assert.Equal(t, "setReadDeadline", (*records)[0].Message)
}

// SetWriteDeadline emits a setWriteDeadline log event.
func TestObservedConnSetWriteDeadlineLogging(t *testing.T) {
	cfg := NewConfig()
	logger, records := newCapturingLogger()

	mockConn := newMinimalConn()
	mockConn.SetWriteDeaFunc = func(time.Time) error { return nil }

	fn := NewObserveConnFunc(cfg, logger)
	observed, _ := fn.Call(context.Background(), mockConn)

	_ = observed.SetWriteDeadline(time.Now().Add(time.Hour))

	require.Len(t, *records, 1)
	assert.Equal(t, "setWriteDeadline", (*records)[0].Message)
}
