package logger

import (
	"bytes"
	"os"
	"testing"
)

func TestSetVerbose(t *testing.T) {
	SetVerbose(true)
	if !verbose {
		t.Error("verbose should be true after SetVerbose(true)")
	}
	SetVerbose(false)
	if verbose {
		t.Error("verbose should be false after SetVerbose(false)")
	}
}

func TestInfo(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = old }()

	Info("hello %s", "world")
	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	if !bytes.Contains(buf.Bytes(), []byte("hello world")) {
		t.Errorf("Info output = %q, want to contain 'hello world'", buf.String())
	}
}
