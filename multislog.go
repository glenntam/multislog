// Package multislog is a custom log/slog logger.
//
// It combines multiple handlers into a single logger.
// It can log to any combination of stderr/logfile/email handlers.
// Individual handlers can have different log levels.
// Log entries can be recorded in a user-specified timezone.
package multislog

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Multislog is a custom logger that has multiple handlers.
// It points to an internal log file that exposes a Close() function.
//
// It's effectively still just a normal *slog.Logger that log/slog can use.
type Multislog struct {
	*slog.Logger

	logFile  *os.File
	timezone *time.Location
	handlers []slog.Handler
}

// Option type to construct a Multislog object with a variable number of options.
type Option func(*Multislog) error

// Close safely closes the log file and any other multihandler resources if they exist.
//
// It is intended to be called as a deferred function at main(), immediately after the logger is instantiated.
// The deferred Close() function ensures the log file is properly closed on normal shutdown and panic unwinding.
// The deferred Close() function won't run on: SIGKILL; power loss; kernel panic; or os.Exit.
//
// See Multislog.New() for usage example.
func (ms *Multislog) Close() {
	// Close handlers first
	for _, h := range ms.handlers {
		c, ok := h.(interface{ Close() error })
		if ok {
			err := c.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "multislog: failed to close handler: %v\n", err)
			}
		}
	}

	// Close log file last
	if ms.logFile != nil {
		err := ms.logFile.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "multislog: failed to close log file: %v\n", err)
		}
		ms.logFile = nil
	}
}

// New is the primary outward Multislog constructor. It is typically called in main().
//
// Example usage (main.go):
//
//	import github.com/glenntam/multislog
//
//	msl := multislog.New(
//	    EnableTimezone("Asia/Hong_Kong"),
//	    EnableConsole(slog.LevelDebug),
//	    EnableLogFile("logfile.json", false, true, slog.LevelDebug),
//	)
//	defer msl.Close()
//	slog.SetDefault(msl.Logger)
//	slog.Info("Logger started...")
//
// By design, New()  panics if any options fail to enable.
func New(opts ...Option) (*Multislog, error) {
	ms := &Multislog{}

	utc := time.UTC
	ms.timezone = utc

	for _, opt := range opts {
		err := opt(ms)
		if err != nil {
			panic(fmt.Errorf("couldn't enable multislog option: %w", err))
		}
	}

	mh := &multihandler{
		handlers: ms.handlers,
		tz:       ms.timezone,
	}
	ms.Logger = slog.New(mh)
	return ms, nil
}

// EnableTimezone forces Multislog to record time stamps in a specific time zone.
// Regardless of which timezone is stamped, all entries are still timezone aware.
//
// timezone argument is any time zone in ISO format. E.g. "America/New_York".
//
// If EnableTimezone is not set, time stamps will be recorded as UTC time zone.
func EnableTimezone(timezone string) Option {
	return func(ms *Multislog) error {
		tz, err := time.LoadLocation(timezone)
		if err != nil {
			return fmt.Errorf("failed to fallback to UTC time zone: %w", err)
		}
		ms.timezone = tz
		return nil
	}
}

// EnableConsole outputs all logs above "level" to stderr.
func EnableConsole(level slog.Level) Option {
	return func(ms *Multislog) error {
		consoleHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
		ms.handlers = append(ms.handlers, consoleHandler)
		return nil
	}
}

// EnableLogFile outputs all logs above "level" to a log file.
//
// allowRead makes the log file world-readable.
// clearOnRestart deletes the existing log file on every run (useful when rapid prototyping).
func EnableLogFile(filename string, allowRead, clearOnRestart bool, level slog.Level) Option {
	return func(ms *Multislog) error {
		file, err := openLogFile(filename, allowRead, clearOnRestart)
		if err != nil {
			return err
		}
		ms.logFile = file
		fileHandler := slog.NewJSONHandler(file, &slog.HandlerOptions{Level: level})
		ms.handlers = append(ms.handlers, fileHandler)
		return nil
	}
}

// EnableEmail outputs all logs above "level" to email.
func EnableEmail(host, port, username, password, sender, recipient string, level slog.Level) Option {
	return func(ms *Multislog) error {
		sc := newSMTPClient(port, host, username, password, sender, recipient)
		emailHandler := newEmailHandler(sc, level)
		ms.handlers = append(ms.handlers, emailHandler)
		return nil
	}
}

var errInvalidLogFileName = errors.New("invalid log file name")

// Helper function for multisloggers to set the log file.
func openLogFile(filename string, allowRead, clearOnRestart bool) (*os.File, error) {
	// Security checks for validity filename
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("resolve executable path: %w", err)
	}

	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return nil, fmt.Errorf("resolve executable symlinks: %w", err)
	}

	baseDir := filepath.Dir(exePath)

	cleanName := filepath.Clean(filename)
	if cleanName != filename || strings.Contains(cleanName, string(os.PathSeparator)) {
		return nil, fmt.Errorf("%w: %q", errInvalidLogFileName, filename)
	}

	logPath := filepath.Join(baseDir, cleanName)

	logPath, err = filepath.EvalSymlinks(logPath)
	if err != nil {
		return nil, fmt.Errorf("resolve log path: %w", err)
	}
	if !strings.HasPrefix(logPath+string(os.PathSeparator), baseDir+string(os.PathSeparator)) {
		return nil, fmt.Errorf("log file escapes executable directory: %w", err)
	}

	// Assemble Log file permissions
	const (
		permOwnerRead = 0o600
		permWorldRead = 0o644
	)

	flags := os.O_CREATE
	logFilePermission := os.FileMode(permOwnerRead)

	if allowRead {
		flags |= os.O_RDWR
		logFilePermission = os.FileMode(permWorldRead)
	} else {
		flags |= os.O_WRONLY
	}

	if clearOnRestart {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_APPEND
	}

	logFile, err := os.OpenFile(logPath, flags, logFilePermission)
	if err != nil {
		return nil, fmt.Errorf("couldn't open %v logfile: %w", filename, err)
	}
	return logFile, nil
}
