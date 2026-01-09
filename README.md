# multislog
A custom GO standard library structured logger that handles multiple handlers (including stderr, log file, email), each with different log levels.

Zero dependencies.

---

## Usage example (in main.go):

```
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

