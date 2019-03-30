// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os/exec"
	"bytes"
	"flag"
	"fmt"
	"os"
	"io"
	"strings"
	"path/filepath"
	"golang.org/x/net/html"
)

var exitCode = 0

func main() {
	vugufmtMain()
	os.Exit(exitCode)
}

func vugufmtMain() {
	// Handle input flags
	flag.Usage = func(){
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
	doc, err := html.Parse(in)
	if err != nil {
		return err
	}
	if err := parseHTML(doc); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse file: %v\n", err)
	}

	// At this point, re-print the parse tree.
	if err := html.Render(out, doc); err != nil {
		return err
	}

	return nil
}

func parseHTML(n *html.Node) error {
	// Are we in the child of a script tag?
	scriptType := ""
	if n.Type == html.TextNode && n.Parent.Type == html.ElementNode && n.Parent.Data == "script" {
		for _, tag := range n.Parent.Attr {
			if tag.Key == "type" {
				scriptType = tag.Val
				break
			}
		}
	}
	// Handle the script type formatting.
	switch scriptType {
	case "application/x-go":
		fmted, err := runGoFmt(n.Data)
		// Exit out on error.
		if err != nil {
			return err
		}
		// Save over Data with the nice version.
		n.Data = fmted
	}
	// Continue on with the recursive pass.
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if err := parseHTML(c); err != nil {
			return err
		}
	}
	return nil
}

func runGoFmt(input string) (string, error) {
	// build up command to run
	cmd := exec.Command("gofmt")

	// I need to capture output
	var fmtOutput bytes.Buffer
	cmd.Stderr = &fmtOutput
	cmd.Stdout = &fmtOutput

	// also set up input pipe
	read, write := io.Pipe()
	cmd.Stdin = read

	// copy down environment variables
	cmd.Env = os.Environ()

	// start gofmt
	if err := cmd.Start(); err != nil {
		return input, fmt.Errorf("can't run gofmt: %s", err)
	}

	// stream in the raw source
	if _, err := write.Write([]byte(input)); err != nil && err != io.ErrClosedPipe {
		return input, fmt.Errorf("gofmt failed: %s", err)
	}

	write.Close()
	// wait until gofmt is done
	if err := cmd.Wait(); err != nil {
		return input, fmt.Errorf("go fmt error %v; full output: %s", err, string(fmtOutput.Bytes()))
	}

	
	return string(fmtOutput.Bytes()), nil
}