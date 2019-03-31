package vugufmt

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVuguFmtNoError(t *testing.T) {
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
		assert.NoError(t, RunFmt(strings.NewReader(testFileString), &buf), f)
		prettyVersion := buf.String()

		// make sure nothing changed!
		assert.NotNil(t, buf.String(), f)
		assert.Equal(t, testFileString, prettyVersion, f)

		ioutil.WriteFile(absPath+".html", []byte(prettyVersion), 0644)
	}

	err := filepath.Walk("./testdata/ok/", func(path string, info os.FileInfo, err error) error {
		fmt(path)
		return nil
	})

	assert.NoError(t, err)
}
