package multislog

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//
// helpers
//

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stderr = w

	fn()

	_ = w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	// #nosec G304 -- path is controlled by test
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	return string(b)
}

func assertPanicsWith(t *testing.T, want error, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		err, ok := r.(error)
		if !ok {
			t.Fatalf("panic value is not error: %T", r)
		}
		if !errors.Is(err, want) {
			t.Fatalf("expected %v, got %v", want, err)
		}
	}()
	fn()
}

//
// construction
//

func TestNew_NoOptions(t *testing.T) {
	ms := New()
	if ms.Logger == nil {
		t.Fatal("logger is nil")
	}
	ms.Close()
}

func TestEnableTimezone_Invalid(t *testing.T) {
	assertPanicsWith(t, errInvalidOption, func() {
		New(EnableTimezone("Not/A_Timezone"))
	})
}

func TestEnableTimezone_Valid(_ *testing.T) {
	ms := New(EnableTimezone("UTC"))
	ms.Close()
}

//
// console
//

func TestEnableConsole_WritesToStderr(t *testing.T) {
	output := captureStderr(t, func() {
		ms := New(EnableConsole(slog.LevelInfo))
		defer ms.Close()

		slog.SetDefault(ms.Logger)
		slog.Info("hello console")
	})
	if !strings.Contains(output, "hello console") {
		t.Fatalf("expected log output, got %q", output)
	}
}

//
// logfile behavior
//

func TestEnableLogFile_CreatesFile(t *testing.T) {
	ms := New(EnableLogFile(slog.LevelInfo, "test.log", false, true))
	defer ms.Close()

	exe, _ := os.Executable()
	logPath := filepath.Join(filepath.Dir(exe), "test.log")

	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("log file not created: %v", err)
	}

	if err := os.Remove(logPath); err != nil && !os.IsNotExist(err) {
		t.Fatalf("remove failed: %v", err)
	}
}

func TestLogFile_AppendVsTruncate(t *testing.T) {
	exe, _ := os.Executable()
	dir := filepath.Dir(exe)
	logPath := filepath.Join(dir, "append.log")

	// first write
	{
		ms := New(EnableLogFile(slog.LevelInfo, "append.log", false, true))
		slog.SetDefault(ms.Logger)
		slog.Info("first")
		ms.Close()
	}

	// append
	{
		ms := New(EnableLogFile(slog.LevelInfo, "append.log", false, false))
		slog.SetDefault(ms.Logger)
		slog.Info("second")
		ms.Close()
	}

	content := readFile(t, logPath)
	if !strings.Contains(content, "first") || !strings.Contains(content, "second") {
		t.Fatalf("append failed, content=%q", content)
	}

	// truncate
	{
		ms := New(EnableLogFile(slog.LevelInfo, "append.log", false, true))
		slog.SetDefault(ms.Logger)
		slog.Info("fresh")
		ms.Close()
	}

	content = readFile(t, logPath)
	if strings.Contains(content, "first") || strings.Contains(content, "second") {
		t.Fatalf("truncate failed, content=%q", content)
	}
	if !strings.Contains(content, "fresh") {
		t.Fatalf("expected fresh entry, got %q", content)
	}

	if err := os.Remove(logPath); err != nil && !os.IsNotExist(err) {
		t.Fatalf("remove failed: %v", err)
	}
}

//
// filename validation / sandboxing
//

func TestEnableLogFile_InvalidFilename(t *testing.T) {
	cases := []string{
		"../evil.log",
		"/absolute.log",
		"subdir/file.log",
		"..",
	}

	for _, name := range cases {
		func(name string) {
			assertPanicsWith(t, errInvalidOption, func() {
				New(EnableLogFile(slog.LevelInfo, name, false, true))
			})
		}(name)
	}
}

//
// handler Close semantics
//

type closingHandler struct {
	closed bool
}

func (h *closingHandler) Enabled(context.Context, slog.Level) bool  { return true }
func (h *closingHandler) Handle(context.Context, slog.Record) error { return nil }
func (h *closingHandler) WithAttrs([]slog.Attr) slog.Handler        { return h }
func (h *closingHandler) WithGroup(string) slog.Handler             { return h }
func (h *closingHandler) Close() error {
	h.closed = true
	return nil
}

func TestClose_ClosesHandlers(t *testing.T) {
	h := &closingHandler{}
	ms := &Multislog{
		Logger:   slog.New(h),
		handlers: []slog.Handler{h},
	}

	ms.Close()
	if !h.closed {
		t.Fatal("expected handler Close() to be called")
	}
}

func TestClose_Idempotent(_ *testing.T) {
	ms := New()
	ms.Close()
	ms.Close() // must not panic
}
