package vugufmt

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

func runGoFmt(input io.Reader, output io.Writer, simplify bool) error {
	// build up command to run
	cmd := exec.Command("gofmt")

	if simplify {
		cmd.Args = []string{"-s"}
	}

	// I need to capture output
	cmd.Stderr = output
	cmd.Stdout = output

	// also set up input pipe
	cmd.Stdin = input

	// copy down environment variables
	cmd.Env = os.Environ()

	// start gofmt
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("can't run gofmt: %s", err)
	}

	// wait until gofmt is done
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("go fmt error %v", err)
	}

	return nil
}
