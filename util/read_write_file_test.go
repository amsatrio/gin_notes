package util

import (
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestWriteToFile_CreateAndAppend(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "dir")

	WriteToFile(dir, "a.txt", "hello", false)
	WriteToFile(dir, "a.txt", " world", false)

	got, err := os.ReadFile(filepath.Join(dir, "a.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello world" {
		t.Errorf("content = %q, want %q", got, "hello world")
	}
}

func TestWriteToFile_Overwrite(t *testing.T) {
	dir := t.TempDir()

	WriteToFile(dir, "b.txt", "first content", true)
	WriteToFile(dir, "b.txt", "second", true)

	got, err := os.ReadFile(filepath.Join(dir, "b.txt"))
	if err != nil {
		t.Fatal(err)
	}
	// overwrite opens without O_TRUNC, so only the leading bytes are replaced
	if string(got) != "secondcontent" {
		t.Errorf("content = %q, want %q", got, "secondcontent")
	}
}

func TestReadFromFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "c.txt"), []byte("some text"), 0600); err != nil {
		t.Fatal(err)
	}

	if got := ReadFromFile(dir, "c.txt"); got != "some text" {
		t.Errorf("ReadFromFile = %q", got)
	}
}

func TestWriteThenRead(t *testing.T) {
	dir := t.TempDir()
	WriteToFile(dir, "d.txt", "round trip", false)
	if got := ReadFromFile(dir, "d.txt"); got != "round trip" {
		t.Errorf("ReadFromFile = %q", got)
	}
}

func TestReadCSV(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.csv")
	if err := os.WriteFile(path, []byte("a,b,c\n1,2,3\n"), 0600); err != nil {
		t.Fatal(err)
	}

	got := ReadCSV(path, ',')
	want := [][]string{{"a", "b", "c"}, {"1", "2", "3"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ReadCSV = %v, want %v", got, want)
	}
}

func TestReadCSV_CustomSeparator(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.csv")
	if err := os.WriteFile(path, []byte("a;b\n1;2\n"), 0600); err != nil {
		t.Fatal(err)
	}

	got := ReadCSV(path, ';')
	want := [][]string{{"a", "b"}, {"1", "2"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ReadCSV = %v, want %v", got, want)
	}
}

func TestRemoveAll(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "to_remove")
	WriteToFile(dir, "x.txt", "bye", false)

	if err := RemoveAll(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("expected dir removed, stat err = %v", err)
	}
}

func TestRemoveAll_NonExistent(t *testing.T) {
	if err := RemoveAll(filepath.Join(t.TempDir(), "missing")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// fatalPanicWriter makes log.Fatal panic (from the log writer) before os.Exit
// is reached, so the failure paths of the file helpers can be exercised safely.
type fatalPanicWriter struct{}

type fatalCalled struct{ msg string }

func (fatalPanicWriter) Write(p []byte) (int, error) {
	panic(fatalCalled{msg: string(p)})
}

// expectFatal runs fn with log output redirected to a panicking writer and
// reports whether log.Fatal (or any log output) was triggered.
func expectFatal(t *testing.T, fn func()) (called bool) {
	t.Helper()
	oldWriter := log.Writer()
	log.SetOutput(fatalPanicWriter{})
	defer log.SetOutput(oldWriter)

	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(fatalCalled); ok {
				called = true
				return
			}
			panic(r)
		}
	}()
	fn()
	return false
}

func TestReadFromFile_MissingFileIsFatal(t *testing.T) {
	dir := t.TempDir()
	if !expectFatal(t, func() { ReadFromFile(dir, "missing.txt") }) {
		t.Error("expected log.Fatal for missing file")
	}
}

func TestReadCSV_MissingFileIsFatal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.csv")
	if !expectFatal(t, func() { ReadCSV(path, ',') }) {
		t.Error("expected log.Fatal for missing csv")
	}
}

func TestReadCSV_MalformedIsFatal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.csv")
	// second record has a different number of fields -> csv.ErrFieldCount
	if err := os.WriteFile(path, []byte("a,b,c\n1,2\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if !expectFatal(t, func() { ReadCSV(path, ',') }) {
		t.Error("expected log.Fatal for malformed csv")
	}
}

func TestWriteToFile_FolderIsFileIsFatal(t *testing.T) {
	base := t.TempDir()
	blocker := filepath.Join(base, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	// MkdirAll fails because a path component is a regular file
	if !expectFatal(t, func() { WriteToFile(blocker, "f.txt", "c", false) }) {
		t.Error("expected log.Fatal when folder cannot be created")
	}
}

func TestWriteToFile_TargetIsDirectoryIsFatal(t *testing.T) {
	base := t.TempDir()
	if err := os.Mkdir(filepath.Join(base, "sub"), 0700); err != nil {
		t.Fatal(err)
	}
	// opening a directory for writing fails (append mode)
	if !expectFatal(t, func() { WriteToFile(base, "sub", "c", false) }) {
		t.Error("expected log.Fatal when target is a directory (append)")
	}
	// ... and in overwrite mode
	if !expectFatal(t, func() { WriteToFile(base, "sub", "c", true) }) {
		t.Error("expected log.Fatal when target is a directory (overwrite)")
	}
}
