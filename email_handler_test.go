package multislog

import (
	"context"
	"errors"
	"log/slog"
	"testing"
)

var errSMTPFailure = errors.New("smtp failure")

type mockEmailHandler struct {
	closed bool
	fail   bool
}

func (m *mockEmailHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (m *mockEmailHandler) Handle(context.Context, slog.Record) error {
	return nil
}

func (m *mockEmailHandler) WithAttrs([]slog.Attr) slog.Handler {
	return m
}

func (m *mockEmailHandler) WithGroup(string) slog.Handler {
	return m
}

func (m *mockEmailHandler) Close() error {
	m.closed = true
	if m.fail {
		return errSMTPFailure
	}
	return nil
}

func TestClose_EmailHandler_CloseCalled(t *testing.T) {
	h := &mockEmailHandler{}

	ms := &Multislog{
		Logger:   slog.New(h),
		handlers: []slog.Handler{h},
	}

	ms.Close()

	if !h.closed {
		t.Fatal("expected email handler Close() to be called")
	}
}

func TestClose_EmailHandler_CloseErrorIgnored(t *testing.T) {
	h1 := &mockEmailHandler{fail: true}
	h2 := &mockEmailHandler{}

	ms := &Multislog{
		Logger:   slog.New(h1),
		handlers: []slog.Handler{h1, h2},
	}

	// Must not panic
	ms.Close()

	if !h1.closed || !h2.closed {
		t.Fatal("expected all handlers to be closed even after error")
	}
}
