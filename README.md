# multislog

[![GoDoc](https://img.shields.io/badge/pkg.go.dev-doc-blue)](http://pkg.go.dev/github.com/glenntam/multislog)
[![Go Report Card](https://goreportcard.com/badge/github.com/glenntam/multislog)](https://goreportcard.com/report/github.com/glenntam/multislog)

Package multislog is a custom log/slog logger.

It combines multiple handlers into a single logger.
It can log to any combination of stderr/logfile/email handlers.
Individual handlers can have different log levels.
Log entries can be recorded in a user-specified timezone.

## Types

### type [Multislog](https://github.com/glenntam/multislog/blob/main/multislog.go#L22)

`type Multislog struct { ... }`

Multislog is a custom logger that has multiple handlers.
It points to an internal log file that exposes a Close() function.

It's effectively still just a normal *slog.Logger that log/slog can use.

#### func [New](https://github.com/glenntam/multislog/blob/main/multislog.go#L68)

`func New(opts ...Option) (*Multislog, error)`

New is the primary outward Multislog constructor. It is typically called in main().

Example usage (main.go):

```go
import github.com/glenntam/multislog

msl, err := multislog.New(
    EnableTimezone("Asia/Hong_Kong"),
    EnableConsole(slog.LevelDebug),
    EnableLogFile("logfile.json", false, true, slog.LevelDebug),
)
if err != nil {
    panic(err.Error())
}
defer msl.Close()
slog.SetDefault(msl.Logger)
slog.Info("Logger started...")
```

#### func (*Multislog) [Close](https://github.com/glenntam/multislog/blob/main/multislog.go#L40)

`func (ms *Multislog) Close()`

Close safely closes the log file if one exists.

It is intended to be called as a deferred function at main(), immediately after the logger is instantiated.
The deferred Close() function ensures the log file is properly closed on normal shutdown and panic unwinding.
The deferred Close() function won't run on: SIGKILL; power loss; kernel panic; or os.Exit.

See Multislog.New() for usage example.

### type [Option](https://github.com/glenntam/multislog/blob/main/multislog.go#L31)

`type Option func(*Multislog) error`

Option type to construct a Multislog object with a variable number of options.

#### func [EnableConsole](https://github.com/glenntam/multislog/blob/main/multislog.go#L107)

`func EnableConsole(level slog.Level) Option`

EnableConsole outputs all logs above "level" to stderr.

#### func [EnableEmail](https://github.com/glenntam/multislog/blob/main/multislog.go#L133)

`func EnableEmail(host, port, username, password, sender, recipient string, level slog.Level) Option`

EnableEmail outputs all logs above "level" to email.

#### func [EnableLogFile](https://github.com/glenntam/multislog/blob/main/multislog.go#L119)

`func EnableLogFile(filename string, allowRead, clearOnRestart bool, level slog.Level) Option`

EnableLogFile outputs all logs above "level" to a log file.

allowRead makes the log file world-readable.
clearOnRestart deletes the existing log file on every run (useful when rapid prototyping).

#### func [EnableTimezone](https://github.com/glenntam/multislog/blob/main/multislog.go#L95)

`func EnableTimezone(timezone string) Option`

EnableTimezone forces Multislog to record time stamps in a specific time zone.
Regardless of which timezone is stamped, all entries are still timezone aware.

timezone argument is any time zone in ISO format. E.g. "America/New_York".

If EnableTimezone is not set, time stamps will be recorded as UTC time zone.
