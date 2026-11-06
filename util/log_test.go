package util

import (
	"bytes"
	"errors"
	"log"
	"os"
	"strings"
	"testing"
)

func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	oldWriter := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(oldWriter) })
	return &buf
}

func TestIsLogged(t *testing.T) {
	tests := []struct {
		mode    string
		level   string
		want    bool
		setMode bool
	}{
		{"ERROR", "ERROR", true, true},
		{"ERROR", "INFO", false, true},
		{"ERROR", "DEBUG", false, true},
		{"INFO", "ERROR", true, true},
		{"INFO", "INFO", true, true},
		{"INFO", "DEBUG", false, true},
		{"DEBUG", "ERROR", true, true},
		{"DEBUG", "INFO", true, true},
		{"DEBUG", "DEBUG", true, true},
		{"", "DEBUG", true, true},
		{"UNKNOWN", "INFO", true, true},
	}
	for _, tt := range tests {
		t.Run(tt.mode+"_"+tt.level, func(t *testing.T) {
			t.Setenv("LOG_MODE", tt.mode)
			if got := IsLogged(tt.level); got != tt.want {
				t.Errorf("LOG_MODE=%q IsLogged(%q) = %v, want %v", tt.mode, tt.level, got, tt.want)
			}
		})
	}
}

func TestIsLogged_EnvUnset(t *testing.T) {
	t.Setenv("LOG_MODE", "")
	if err := os.Unsetenv("LOG_MODE"); err != nil {
		t.Fatal(err)
	}
	if !IsLogged("DEBUG") {
		t.Error("expected everything logged when LOG_MODE unset")
	}
}

func TestLog(t *testing.T) {
	t.Setenv("LOG_MODE", "DEBUG")
	buf := captureLog(t)

	Log("INFO", "pkg", "fn", "hello")
	out := buf.String()
	if !strings.Contains(out, "[INFO] => pkg > fn > hello") {
		t.Errorf("unexpected log output: %q", out)
	}
}

func TestLog_Suppressed(t *testing.T) {
	t.Setenv("LOG_MODE", "ERROR")
	buf := captureLog(t)

	Log("INFO", "pkg", "fn", "hidden")
	if buf.Len() != 0 {
		t.Errorf("expected no output, got %q", buf.String())
	}

	Log("ERROR", "pkg", "fn", "shown")
	if !strings.Contains(buf.String(), "[ERROR] => pkg > fn > shown") {
		t.Errorf("unexpected log output: %q", buf.String())
	}
}

func TestLogError(t *testing.T) {
	t.Setenv("LOG_MODE", "INFO")
	buf := captureLog(t)

	LogError("pkg", "fn", "failed", errors.New("boom"))
	out := buf.String()
	for _, want := range []string{"[ERROR] => pkg > fn > failed", "Error: boom", "Stack Trace:"} {
		if !strings.Contains(out, want) {
			t.Errorf("log output missing %q: %q", want, out)
		}
	}
}

func TestLogError_NilError(t *testing.T) {
	t.Setenv("LOG_MODE", "ERROR")
	buf := captureLog(t)

	LogError("pkg", "fn", "no error", nil)
	out := buf.String()
	if !strings.Contains(out, "[ERROR] => pkg > fn > no error") {
		t.Errorf("unexpected log output: %q", out)
	}
	if strings.Contains(out, "Stack Trace") {
		t.Errorf("did not expect stack trace: %q", out)
	}
}

func TestLogError_Suppressed(t *testing.T) {
	// ERROR level is always logged for known modes; unset mode logs everything.
	// There is no mode that suppresses ERROR, so just verify it is emitted in all modes.
	for _, mode := range []string{"ERROR", "INFO", "DEBUG"} {
		t.Setenv("LOG_MODE", mode)
		buf := captureLog(t)
		LogError("p", "f", "m", nil)
		if buf.Len() == 0 {
			t.Errorf("expected output for mode %s", mode)
		}
	}
}

func TestLogAPI(t *testing.T) {
	t.Setenv("LOG_MODE", "INFO")
	buf := captureLog(t)

	LogAPI("127.0.0.1", "/v1/notes", 200, "5ms")
	out := buf.String()
	want := "[INFO] => API > IP: 127.0.0.1, requested path: /v1/notes, status: 200, elapsed time: 5ms"
	if !strings.Contains(out, want) {
		t.Errorf("unexpected log output: %q", out)
	}
}

func TestLogAPI_Suppressed(t *testing.T) {
	t.Setenv("LOG_MODE", "ERROR")
	buf := captureLog(t)

	LogAPI("127.0.0.1", "/v1/notes", 200, "5ms")
	if buf.Len() != 0 {
		t.Errorf("expected no output, got %q", buf.String())
	}
}
