package vugufmt

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptsCustom(t *testing.T) {
	jsFormat := func(f *Formatter) {
		f.FmtScripts["js"] = func(input io.Reader, output io.Writer) error {
			return nil
		}
	}
	formatter := NewFormatter(jsFormat)
	assert.NotNil(t, formatter.FmtScripts["js"])
}

func TestOptsGoFmt(t *testing.T) {
	gofmt := UseGoFmt(false)
	formatter := NewFormatter(gofmt)
	assert.NotNil(t, formatter.FmtScripts["application/x-go"])
}

func TestOptsGoFmtSimple(t *testing.T) {
	gofmt := UseGoFmt(true)
	formatter := NewFormatter(gofmt)
	assert.NotNil(t, formatter.FmtScripts["application/x-go"])
}

func TestOptsGoImports(t *testing.T) {
	goimports := UseGoImports
	formatter := NewFormatter(goimports)
	assert.NotNil(t, goimports, formatter.FmtScripts["application/x-go"])
}
