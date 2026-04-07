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
