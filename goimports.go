package vugufmt

import (
	"bytes"
	"errors"
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

	var resBuff bytes.Buffer

	// I need to capture output
	cmd.Stderr = &resBuff
	cmd.Stdout = &resBuff

	// also set up input pipe
	cmd.Stdin = input

	// copy down environment variables
	cmd.Env = os.Environ()

	// start gofmt
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("can't run goimports: %s", err)
	}

	// wait until gofmt is done
	err := cmd.Wait()

	// Get all the output
	res := resBuff.Bytes()
	// Send the output to the output buffer
	io.Copy(output, bytes.NewReader(res))
	// Wrap the output in an error.
	if err != nil {
		return errors.New(string(res))
	}

	return nil
}
