// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file is a modification of https://golang.org/src/cmd/gofmt/gofmt.go

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/erinpentecost/vugufmt"
)

var (
	exitCode    = 0
	list        = flag.Bool("l", false, "list files whose formatting differs from vugufmt's")
	write       = flag.Bool("w", false, "write result to (source) file instead of stdout")
	simplifyAST = flag.Bool("s", false, "simplify code")
	doDiff      = flag.Bool("d", false, "display diffs instead of rewriting files")
)

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
		if err := processFile("<standard input>", os.Stdin, os.Stdout); err != nil {
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
			if err := processFile(path, nil, os.Stdout); err != nil {
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
		err = processFile(path, nil, os.Stdout)
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

func processFile(filename string, in io.Reader, out io.Writer) error {
	var perm os.FileMode = 0644
	// open the file if needed
	if in == nil {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			return err
		}
		in = f
		perm = fi.Mode().Perm()
	}

	src, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}

	var resBuff bytes.Buffer

	formatter := vugufmt.NewFormatter(vugufmt.UseGoFmt(*simplifyAST))

	if err := formatter.Format(bytes.NewReader(src), &resBuff); err != nil {
		return err
	}
	res := resBuff.Bytes()

	if !bytes.Equal(src, res) {
		// formatting has changed
		if *list {
			fmt.Fprintln(out, filename)
		}

		if *write {
			// make a temporary backup before overwriting original
			bakname, err := backupFile(filename+".", src, perm)
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(filename, res, perm)
			if err != nil {
				os.Rename(bakname, filename)
				return err
			}
			err = os.Remove(bakname)
			if err != nil {
				return err
			}
		}

		if *doDiff {
			data, err := diff(src, res, filename)
			if err != nil {
				return fmt.Errorf("computing diff: %s", err)
			}
			fmt.Printf("diff -u %s %s\n", filepath.ToSlash(filename+".orig"), filepath.ToSlash(filename))
			out.Write(data)
		}
	}

	if !*list && !*write && !*doDiff {
		_, err = out.Write(res)
	}
	return err
}

func writeTempFile(dir, prefix string, data []byte) (string, error) {
	file, err := ioutil.TempFile(dir, prefix)
	if err != nil {
		return "", err
	}
	_, err = file.Write(data)
	if err1 := file.Close(); err == nil {
		err = err1
	}
	if err != nil {
		os.Remove(file.Name())
		return "", err
	}
	return file.Name(), nil
}

func diff(b1, b2 []byte, filename string) (data []byte, err error) {
	f1, err := writeTempFile("", "vugufmt", b1)
	if err != nil {
		return
	}
	defer os.Remove(f1)
	f2, err := writeTempFile("", "vugufmt", b2)
	if err != nil {
		return
	}
	defer os.Remove(f2)
	cmd := "diff"
	if runtime.GOOS == "plan9" {
		cmd = "/bin/ape/diff"
	}
	data, err = exec.Command(cmd, "-u", f1, f2).CombinedOutput()
	if len(data) > 0 {
		// diff exits with a non-zero status when the files don't match.
		// Ignore that failure as long as we get output.
		return replaceTempFilename(data, filename)
	}
	return
}

// replaceTempFilename replaces temporary filenames in diff with actual one.
func replaceTempFilename(diff []byte, filename string) ([]byte, error) {

	bs := bytes.SplitN(diff, []byte{'\n'}, 3)

	if len(bs) < 3 {
		return nil, fmt.Errorf("got unexpected diff for %s", filename)
	}

	// Preserve timestamps.
	var t0, t1 []byte

	if i := bytes.LastIndexByte(bs[0], '\t'); i != -1 {
		t0 = bs[0][i:]
	}

	if i := bytes.LastIndexByte(bs[1], '\t'); i != -1 {
		t1 = bs[1][i:]
	}

	// Always print filepath with slash separator.
	f := filepath.ToSlash(filename)
	bs[0] = []byte(fmt.Sprintf("--- %s%s", f+".orig", t0))
	bs[1] = []byte(fmt.Sprintf("+++ %s%s", f, t1))
	return bytes.Join(bs, []byte{'\n'}), nil
}

const chmodSupported = runtime.GOOS != "windows"

// backupFile writes data to a new file named filename<number> with permissions perm,
// with <number randomly chosen such that the file name is unique. backupFile returns
// the chosen file name.
func backupFile(filename string, data []byte, perm os.FileMode) (string, error) {

	// create backup file
	f, err := ioutil.TempFile(filepath.Dir(filename), filepath.Base(filename))
	if err != nil {
		return "", err
	}

	bakname := f.Name()

	if chmodSupported {
		err = f.Chmod(perm)
		if err != nil {
			f.Close()
			os.Remove(bakname)
			return bakname, err
		}
	}

	// write data to backup file
	_, err = f.Write(data)
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return bakname, err
}
