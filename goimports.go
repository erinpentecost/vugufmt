package vugufmt

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

// UseGoImports sets the formatter to use goimports on x-go blocks.
func UseGoImports(f *Formatter) {
	f.FmtScripts["application/x-go"] = func(input io.Reader, output io.Writer) error {
		return runGoImports(input, output)
	}
}

func runGoImports(input io.Reader, output io.Writer) error {
	// build up command to run
	cmd := exec.Command("goimports")

	// I need to capture output
	cmd.Stderr = output
	cmd.Stdout = output

	// also set up input pipe
	cmd.Stdin = input

	// copy down environment variables
	cmd.Env = os.Environ()

	// start goimports
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("can't run goimports: %s", err)
	}

	// wait until goimports is done
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("goimports error %v", err)
	}

	return nil
}
