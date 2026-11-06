package initializer

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func restoreLogOutput(t *testing.T) {
	t.Helper()
	prevWriter := log.Writer()
	t.Cleanup(func() { log.SetOutput(prevWriter) })
}

func todayLogFile() string {
	return filepath.Join("log", "log_"+time.Now().Format("2006-01-02")+".log")
}

func TestLoggerInitCreatesLogDir(t *testing.T) {
	restoreLogOutput(t)
	t.Chdir(t.TempDir())
	t.Setenv("LOG_CLEAR", "false")

	LoggerInit()
	log.SetOutput(io.Discard) // release the file for subsequent assertions

	data, err := os.ReadFile(todayLogFile())
	if err != nil {
		t.Fatalf("log file not created: %v", err)
	}
	if !strings.Contains(string(data), "hi there") {
		t.Errorf("log file missing ASCII art, got: %q", string(data))
	}
}

func TestLoggerInitExistingDirNoClear(t *testing.T) {
	restoreLogOutput(t)
	t.Chdir(t.TempDir())
	t.Setenv("LOG_CLEAR", "false")

	if err := os.Mkdir("log", os.ModePerm); err != nil {
		t.Fatal(err)
	}
	keep := filepath.Join("log", "keep.txt")
	if err := os.WriteFile(keep, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	LoggerInit()
	log.SetOutput(io.Discard)

	if _, err := os.Stat(keep); err != nil {
		t.Errorf("existing file should be kept: %v", err)
	}
	if _, err := os.Stat(todayLogFile()); err != nil {
		t.Errorf("log file not created: %v", err)
	}
}

func TestLoggerInitClear(t *testing.T) {
	restoreLogOutput(t)
	t.Chdir(t.TempDir())
	t.Setenv("LOG_CLEAR", "true")

	if err := os.Mkdir("log", os.ModePerm); err != nil {
		t.Fatal(err)
	}
	old := filepath.Join("log", "old.log")
	if err := os.WriteFile(old, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	LoggerInit()
	log.SetOutput(io.Discard)

	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Errorf("old log file should be removed, stat err = %v", err)
	}
	if _, err := os.Stat(todayLogFile()); err != nil {
		t.Errorf("new log file not created: %v", err)
	}
}
