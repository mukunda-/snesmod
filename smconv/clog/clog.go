// SNESMOD
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

// Console Log
package clog

import (
	"fmt"
	"os"
)

func Infoln(a ...any) {
	fmt.Fprintln(os.Stderr, append([]any{"INFO"}, a...)...)
}

func Errorln(a ...any) {
	fmt.Fprintln(os.Stderr, append([]any{"ERR "}, a...)...)
}

func Errorf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "ERR "+format, a...)
}
