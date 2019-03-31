package vugufmt

import (
	"io"
)

// RunFmt runs vugufmt on input, and sends a pretty version of
// if out to output. If there is an error, throw away output!
func RunFmt(input io.Reader, output io.Writer) error {
	return runHTMLFmt(input, output)
}
