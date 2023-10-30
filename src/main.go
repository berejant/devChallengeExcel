//go:build !test

package main

import (
	"os"
)

func main() {
	HandleExitError(os.Stderr, RunApp())
}
