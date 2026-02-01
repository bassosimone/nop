package nop

import (
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// httpBodyWrap wraps an HTTP body so that we emit structured log events
// lazily: httpBodyStreamStart on the first Read, and httpBodyStreamDone
// on Close (only if at least one Read happened).
func httpBodyWrap(
	body io.ReadCloser,
	errClass ErrClassifier,
	laddr string,
	logger SLogger,
	protocol string,
	raddr string,
	timeNow func() time.Time,
) io.ReadCloser {
	return &httpBodyWrapper{
		body:      body,
		closeOnce: sync.Once{},
		didRead:   atomic.Bool{},
		errClass:  errClass,
		laddr:     laddr,
		logger:    logger,
		protocol:  protocol,
		raddr:     raddr,
		readOnce:  sync.Once{},
		timeNow:   timeNow,
		t0:        time.Time{},
	}
}

type httpBodyWrapper struct {
	// body is the actual body.
	body io.ReadCloser

	// didRead tracks whether at least one Read happened.
	didRead atomic.Bool

	// errClass is the err classifier in use.
	errClass ErrClassifier

	// laddr is the local address.
	laddr string

	// logger is the [SLogger] in use.
	logger SLogger

	// closeOnce ensures that Close has "once" semantics.
	closeOnce sync.Once

	// protocol is the network protocol ("tcp" or "udp").
	protocol string

	// raddr is the remote address.
	raddr string

	// readOnce ensures we log httpBodyStreamStart only once.
	readOnce sync.Once

	// t0 is the time when we started reading the body.
	t0 time.Time

	// timeNow mocks [time.Now].
	timeNow func() time.Time
}

var _ io.ReadCloser = &httpBodyWrapper{}

// Close implements [io.ReadCloser].
func (b *httpBodyWrapper) Close() (err error) {
	b.closeOnce.Do(func() {
		err = b.body.Close()
		if b.didRead.Load() { // acquire: t0 is visible if this returns true
			b.logger.Info(
				"httpBodyStreamDone",
				slog.Any("err", err),
				slog.String("errClass", b.errClass.Classify(err)),
				slog.String("localAddr", b.laddr),
				slog.String("protocol", b.protocol),
				slog.String("remoteAddr", b.raddr),
				slog.Time("t0", b.t0),
				slog.Time("t", b.timeNow()),
			)
		}
	})
	return
}

// Read implements [io.ReadCloser].
func (b *httpBodyWrapper) Read(buffer []byte) (int, error) {
	b.readOnce.Do(func() {
		b.t0 = b.timeNow()    // write t0 BEFORE the atomic store (release)
		b.didRead.Store(true) // release: makes t0 visible to Close
		b.logger.Info(
			"httpBodyStreamStart",
			slog.String("localAddr", b.laddr),
			slog.String("protocol", b.protocol),
			slog.String("remoteAddr", b.raddr),
			slog.Time("t", b.t0),
		)
	})
	return b.body.Read(buffer)
}
