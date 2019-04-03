package vugufmt

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// UseGoImports sets the formatter to use goimports on x-go blocks.
func UseGoImports(f *Formatter) {
	f.ScriptFormatters["application/x-go"] = func(input []byte) ([]byte, error) {
		return runGoImports(input)
	}
}

func runGoImports(input []byte) ([]byte, error) {
	// build up command to run
	cmd := exec.Command("goimports")

	var resBuff bytes.Buffer

	// I need to capture output
	cmd.Stderr = &resBuff
	cmd.Stdout = &resBuff

	// also set up input pipe
	cmd.Stdin = bytes.NewReader(input)

	// copy down environment variables
	cmd.Env = os.Environ()

	// start gofmt
	if err := cmd.Start(); err != nil {
		return input, fmt.Errorf("can't run goimports: %s", err)
	}

	// wait until gofmt is done
	err := cmd.Wait()

	// Get all the output
	res := resBuff.Bytes()

	// Wrap the output in an error.
	if err != nil {
		return input, errors.New(string(res))
	}

	return res, nil
}
