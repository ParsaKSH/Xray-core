package dispatcher

import (
	"context"

	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/common/buf"
	"github.com/xtls/xray-core/common/ratelimit"
	"github.com/xtls/xray-core/features/stats"
)

type SizeStatWriter struct {
	Counter stats.Counter
	Writer  buf.Writer
}

func (w *SizeStatWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	w.Counter.Add(int64(mb.Len()))
	return w.Writer.WriteMultiBuffer(mb)
}

func (w *SizeStatWriter) Close() error {
	return common.Close(w.Writer)
}

func (w *SizeStatWriter) Interrupt() {
	common.Interrupt(w.Writer)
}

// RateLimitWriter wraps a buf.Writer with bandwidth rate limiting.
// It optionally also counts bytes for stats.
type RateLimitWriter struct {
	Counter stats.Counter
	Writer  buf.Writer
	Limiter *ratelimit.Limiter
	Ctx     context.Context
}

func (w *RateLimitWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	n := int(mb.Len())
	if w.Limiter != nil && n > 0 {
		if err := w.Limiter.Wait(w.Ctx, n); err != nil {
			buf.ReleaseMulti(mb)
			return err
		}
	}
	if w.Counter != nil {
		w.Counter.Add(int64(n))
	}
	return w.Writer.WriteMultiBuffer(mb)
}

func (w *RateLimitWriter) Close() error {
	return common.Close(w.Writer)
}

func (w *RateLimitWriter) Interrupt() {
	common.Interrupt(w.Writer)
}

// RateLimitReader wraps a buf.Reader with bandwidth rate limiting.
// This is needed for protocols (like SS2022 via singbridge) that read
// from link.Reader directly instead of going through a Writer wrapper.
type RateLimitReader struct {
	Reader  buf.Reader
	Limiter *ratelimit.Limiter
	Ctx     context.Context
}

func (r *RateLimitReader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	mb, err := r.Reader.ReadMultiBuffer()
	if err != nil {
		return mb, err
	}
	n := int(mb.Len())
	if r.Limiter != nil && n > 0 {
		if waitErr := r.Limiter.Wait(r.Ctx, n); waitErr != nil {
			buf.ReleaseMulti(mb)
			return nil, waitErr
		}
	}
	return mb, nil
}

func (r *RateLimitReader) Close() error {
	return common.Close(r.Reader)
}

func (r *RateLimitReader) Interrupt() {
	common.Interrupt(r.Reader)
}
