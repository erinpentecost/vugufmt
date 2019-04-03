package vugufmt

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// UseGoFmt sets the formatter to use gofmt on x-go blocks.
// Set simplifyAST to true to simplify the AST. This is false
// by default for gofmt, and is the same as passing in -s for it.
func UseGoFmt(simplifyAST bool) func(*Formatter) {

	return func(f *Formatter) {
		f.ScriptFormatters["application/x-go"] = func(input []byte) ([]byte, error) {
			return runGoFmt(input, simplifyAST)
		}
	}
}

func runGoFmt(input []byte, simplify bool) ([]byte, error) {
	// build up command to run
	cmd := exec.Command("gofmt")

	if simplify {
		cmd.Args = []string{"-s"}
	}

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
		return input, fmt.Errorf("can't run gofmt: %s", err)
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
