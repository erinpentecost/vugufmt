// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/erinpentecost/vugufmt"
)

var exitCode = 0

func main() {
	vugufmtMain()
	os.Exit(exitCode)
}

func vugufmtMain() {
	// Handle input flags
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: vugufmt [flags] [path ...]\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	// If no file paths given, we are reading from stdin.
	if flag.NArg() == 0 {
		if err := processFile("<standard input>", os.Stdin, os.Stdout, true); err != nil {
			report(err)
		}
		return
	}

	// Otherwise, we need to read a bunch of files
	for i := 0; i < flag.NArg(); i++ {
		path := flag.Arg(i)
		switch dir, err := os.Stat(path); {
		case err != nil:
			report(err)
		case dir.IsDir():
			walkDir(path)
		default:
			if err := processFile(path, nil, os.Stdout, false); err != nil {
				report(err)
			}
		}
	}
}

func walkDir(path string) {
	filepath.Walk(path, visitFile)
}

func visitFile(path string, f os.FileInfo, err error) error {
	if err == nil && isVuguFile(f) {
		err = processFile(path, nil, os.Stdout, false)
	}

	// Don't complain if a file was deleted in the meantime (i.e.
	// the directory changed concurrently while running gofmt).
	if err != nil && !os.IsNotExist(err) {
		report(err)
	}
	return nil
}

func isVuguFile(f os.FileInfo) bool {
	// ignore non-Vugu files
	name := f.Name()
	return !f.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".vugu")
}

func report(err error) {
	fmt.Fprintf(os.Stderr, err.Error())
	exitCode = 2
}

func processFile(filename string, in io.Reader, out io.Writer, stdin bool) error {
	// open the file if needed
	if in == nil {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := f.Stat(); err != nil {
			return err
		}

		in = f
	}

	// First process the html bits
	if err := vugufmt.RunFmt(in, out); err != nil {
		return err
	}

	return nil
}
