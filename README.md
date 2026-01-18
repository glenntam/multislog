# multislog

[![GoDoc](https://img.shields.io/badge/pkg.go.dev-doc-blue)](http://pkg.go.dev/github.com/glenntam/multislog)
[![Go Report Card](https://goreportcard.com/badge/github.com/glenntam/multislog)](https://goreportcard.com/report/github.com/glenntam/multislog)

Package multislog is a custom multilogger that plays nice with Go standard library log/slog.

It can log to console, a log file and email at the same time, each with a different log level.

It is slog-compliant: Anywhere slog is used you can use multislog without having to change any existing code.

For convenience in small projects, log entries can optionally be recorded in a user-specified timezone.

## Types

### type [Multislog](https://github.com/glenntam/multislog/blob/main/multislog.go#L29)

`type Multislog struct { ... }`

Multislog is a custom logger that has multiple handlers.
It points to an internal log file that exposes a Close() function.

It can be used by standard library log/slog and is a slog.Logger in all ways.

#### func [New](https://github.com/glenntam/multislog/blob/main/multislog.go#L91)

`func New(opts ...Option) *Multislog`

New is the primary Multislog constructor. It is typically called in main().

Example usage (main.go):

```go
import github.com/glenntam/multislog

msl := multislog.New(
    EnableTimezone("America/New_York"),
    EnableConsole(slog.LevelDebug),
    EnableLogFile(slog.LevelInfo, "logfile.json", false, true),
    EnableEmail(slog.LevelWarn, "smtp.gmail.com", "465", "admin", "myPassword", "from@gmail.com", "to@email.com"),
)
defer msl.Close()
slog.SetDefault(msl.Logger)
slog.Info("Logger started...")
```

By design, New() panics if any options fail to enable at start.

#### func (*Multislog) [Close](https://github.com/glenntam/multislog/blob/main/multislog.go#L52)

`func (ms *Multislog) Close()`

Close safely closes the log file and any other multihandler resources if they exist.

It is intended to be called as a deferred function at main(), immediately after the logger is instantiated.
The deferred Close() function ensures the log file is properly closed on normal shutdown and panic unwinding.
The deferred Close() function won't run on: SIGKILL; power loss; kernel panic; or os.Exit.

Example (main.go):

```go
msl := multislog.New(EnableConsole(slog.LevelDebug))
defer msl.Close()
```

See multislog.New() for complete usage example.

### type [Option](https://github.com/glenntam/multislog/blob/main/multislog.go#L38)

`type Option func(*Multislog) error`

Option type to construct a Multislog object with a variable number of options.

#### func [EnableConsole](https://github.com/glenntam/multislog/blob/main/multislog.go#L130)

`func EnableConsole(level slog.Level) Option`

EnableConsole outputs all logs above "level" to stderr.

#### func [EnableEmail](https://github.com/glenntam/multislog/blob/main/multislog.go#L156)

`func EnableEmail(level slog.Level, host, port, username, password, sender, recipient string) Option`

EnableEmail outputs all logs above "level" to email.

#### func [EnableLogFile](https://github.com/glenntam/multislog/blob/main/multislog.go#L142)

`func EnableLogFile(level slog.Level, filename string, allowRead, clearOnRestart bool) Option`

EnableLogFile outputs all logs above "level" to a log file.

allowRead makes the log file world-readable.
clearOnRestart deletes the existing log file on every run (useful when rapid prototyping).

#### func [EnableTimezone](https://github.com/glenntam/multislog/blob/main/multislog.go#L118)

`func EnableTimezone(timezone string) Option`

EnableTimezone forces Multislog to record time stamps in a specific time zone.
Regardless of which timezone is stamped, all entries are still timezone aware.

timezone argument is any time zone in ISO format. E.g. "America/New_York".

If EnableTimezone is not set, time stamps will be recorded as UTC time zone.
