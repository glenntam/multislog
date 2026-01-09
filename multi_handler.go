package multislog

import (
	"context"
	"log/slog"
	"time"
)

// Multihandler is a slice of slog.Handlers. It shadows some slog.Handler
// methods to ensure relevant log messages are sent to different handlers,
// since each handler may have different log levels.
//
// Multislog uses a single Multihandler object create a new custom logger.
type multihandler struct {
	tz       *time.Location
	handlers []slog.Handler
}

// Enabled determines if a slog message will be processed.
func (mh *multihandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range mh.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle determines how a slog message will be processed.
// It also overwrites the recorded timezone with the chosen one.
func (mh *multihandler) Handle(ctx context.Context, r slog.Record) error {
	if mh.tz != nil {
		r.Time = r.Time.In(mh.tz)
	}
	for _, h := range mh.handlers {
		if h.Enabled(ctx, r.Level) {
			_ = h.Handle(ctx, r)
		}
	}
	return nil
}

// WithAttrs satisfies handler interface.
func (mh *multihandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	hs := make([]slog.Handler, len(mh.handlers))
	for i, h := range mh.handlers {
		hs[i] = h.WithAttrs(attrs)
	}
	return &multihandler{
		handlers: hs,
		tz:       mh.tz,
	}
}

// WithGroup satisfies handler interface.
func (mh *multihandler) WithGroup(name string) slog.Handler {
	hs := make([]slog.Handler, len(mh.handlers))
	for i, h := range mh.handlers {
		hs[i] = h.WithGroup(name)
	}
	return &multihandler{
		handlers: hs,
		tz:       mh.tz,
	}
}
