package vugufmt

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// UseGoFmt sets the formatter to use gofmt on x-go blocks.
// Set simplifyAST to true to simplify the AST. This is false
// by default for gofmt, and is the same as passing in -s for it.
func UseGoFmt(simplifyAST bool) func(*Formatter) {

	return func(f *Formatter) {
		f.FmtScripts["application/x-go"] = func(input io.Reader, output io.Writer) error {
			return runGoFmt(input, output, simplifyAST)
		}
	}
}

func runGoFmt(input io.Reader, output io.Writer, simplify bool) error {
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
	cmd.Stdin = input

	// copy down environment variables
	cmd.Env = os.Environ()

	// start gofmt
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("can't run gofmt: %s", err)
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
