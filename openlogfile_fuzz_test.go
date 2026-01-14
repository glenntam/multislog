package multislog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func FuzzOpenLogFile(f *testing.F) {
	seeds := []string{
		"test.log",
		"",
		".",
		"..",
		"../evil.log",
		"..\\evil.log",
		"/absolute.log",
		"subdir/file.log",
		"subdir\\file.log",
		"\x00",
		"🔥.log",
	}

	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, filename string) {
		file, err := openLogFile(filename, false, true)
		if err != nil {
			return
		}

		defer func() {
			if err := file.Close(); err != nil {
				t.Fatalf("file.Close failed: %v", err)
			}
		}()

		defer func() {
			if err := os.Remove(file.Name()); err != nil && !os.IsNotExist(err) {
				t.Fatalf("os.Remove failed: %v", err)
			}
		}()

		exe, err := os.Executable()
		if err != nil {
			t.Fatalf("os.Executable failed: %v", err)
		}

		exe, err = filepath.EvalSymlinks(exe)
		if err != nil {
			t.Fatalf("EvalSymlinks failed: %v", err)
		}

		baseDir := filepath.Dir(exe)
		logPath, err := filepath.EvalSymlinks(file.Name())
		if err != nil {
			t.Fatalf("EvalSymlinks failed: %v", err)
		}

		baseDirSep := baseDir + string(os.PathSeparator)
		logPathSep := logPath + string(os.PathSeparator)

		if !strings.HasPrefix(logPathSep, baseDirSep) {
			t.Fatalf("log file escaped executable dir\nbase=%q\nlog=%q\nfilename=%q", baseDir, logPath, filename)
		}
	})
}
