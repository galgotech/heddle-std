package io

import (
	"bytes"
	"errors"
	goio "io"
	"os"
	"testing"
)

// capture runs fn with *stream redirected to a pipe and returns everything
// written to it.
func capture(t *testing.T, stream **os.File, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	original := *stream
	*stream = w
	defer func() { *stream = original }()

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = goio.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close pipe: %v", err)
	}
	out := <-done
	if err := r.Close(); err != nil {
		t.Fatalf("close reader: %v", err)
	}
	return out
}

func TestPrint(t *testing.T) {
	var err error

	t.Run("string", func(t *testing.T) {
		got := capture(t, &os.Stdout, func() { err = Print("hello") })
		if err != nil || got != "hello\n" {
			t.Errorf("wrote %q (err %v), expected %q", got, err, "hello\n")
		}
	})
	t.Run("empty string still prints a newline", func(t *testing.T) {
		got := capture(t, &os.Stdout, func() { err = Print("") })
		if err != nil || got != "\n" {
			t.Errorf("wrote %q (err %v), expected %q", got, err, "\n")
		}
	})
}

// TestPrintReportsWriteFailure pins that Print is genuinely fallible: writing
// to a closed pipe fails, and a Heddle handler is entitled to see it. This is
// also what justifies `? print_error()` on io.print calls in the corpus —
// `?` on an infallible call is rejected (docs/heddle.md §3.5).
func TestPrintReportsWriteFailure(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close reader: %v", err)
	}
	original := os.Stdout
	os.Stdout = w
	defer func() {
		os.Stdout = original
		_ = w.Close()
	}()

	if err := Print("into a closed pipe"); err == nil {
		t.Error("expected a write error, got nil")
	}
}

func TestPrintError(t *testing.T) {
	var err error
	got := capture(t, &os.Stderr, func() {
		err = PrintError(errors.New("boom"))
	})
	if err != nil || got != "boom\n" {
		t.Errorf("wrote %q (err %v), expected %q", got, err, "boom\n")
	}
}

// TestPrintErrorReportsWriteFailure pins that reporting a failure can itself
// fail — which is why a handler body may chain `? log_to_stderr()` on its own
// io.print_error call (docs/heddle.md §3.3).
func TestPrintErrorReportsWriteFailure(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close reader: %v", err)
	}
	original := os.Stderr
	os.Stderr = w
	defer func() {
		os.Stderr = original
		_ = w.Close()
	}()

	if err := PrintError(errors.New("x")); err == nil {
		t.Error("expected a write error, got nil")
	}
}
