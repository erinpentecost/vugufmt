package vugufmt

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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

func TestVuguFmtNoError(t *testing.T) {
	formatter := NewFormatter(UseGoFmt(false))
	fmt := func(f string) {
		// Need to un-relativize the paths
		absPath, err := filepath.Abs(f)

		if filepath.Ext(absPath) != ".vugu" {
			return
		}

		assert.NoError(t, err, f)
		// get a handle on the file
		testFile, err := ioutil.ReadFile(absPath)
		testFileString := string(testFile)
		assert.NoError(t, err, f)
		// run gofmt on it
		var buf bytes.Buffer
		assert.NoError(t, formatter.Format("", strings.NewReader(testFileString), &buf), f)
		prettyVersion := buf.String()

		// make sure nothing changed!
		assert.NotNil(t, buf.String(), f)
		assert.Equal(t, testFileString, prettyVersion, f)

		//ioutil.WriteFile(absPath+".html", []byte(prettyVersion), 0644)
	}

	err := filepath.Walk("./testdata/ok/", func(path string, info os.FileInfo, err error) error {
		fmt(path)
		return nil
	})

	assert.NoError(t, err)
}

func TestEscaping(t *testing.T) {
	// I'd like the > to not be escaped into &gt;
	testCode := "<div vg-if='len(data.bpi.BPI) > 0'></div>"
	formatter := NewFormatter(UseGoFmt(false))
	// run gofmt on it
	var buf bytes.Buffer
	assert.NoError(t, formatter.Format("", strings.NewReader(testCode), &buf), testCode)
	prettyVersion := buf.String()
	assert.Equal(t, testCode, prettyVersion)
}
