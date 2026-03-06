package logger

import (
	"fmt"
	"os"
)

var verbose bool

// SetVerbose enables or disables verbose output.
func SetVerbose(v bool) {
	verbose = v
}

// Info prints an informational message.
func Info(format string, args ...any) {
	fmt.Fprintf(os.Stdout, format+"\n", args...)
}

// Verbose prints a message only when verbose mode is enabled.
func Verbose(format string, args ...any) {
	if verbose {
		fmt.Fprintf(os.Stdout, "[verbose] "+format+"\n", args...)
	}
}

// Error prints an error message to stderr.
func Error(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
}
