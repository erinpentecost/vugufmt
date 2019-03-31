package vugufmt

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/erinpentecost/vugufmt/htmlx"
)

// Formatter allows you to format vugu files.
type Formatter struct {
	// FmtScripts maps script blocks to formatting
	// functions.
	// For each type of script block,
	// we can run it through the supplied function.
	// If the function returns error, we should
	// not accept the output written to the writer.
	// You can add your own custom one for JS, for
	// example. If you want to use gofmt or goimports,
	// see how to apply options in NewFormatter.
	FmtScripts map[string](func(io.Reader, io.Writer) error)
}

// NewFormatter creates a new formatter.
// Pass in vugufmt.UseGoFmt to use gofmt.
// Pass in vugufmt.UseGoImports to use goimports.
func NewFormatter(opts ...func(*Formatter)) *Formatter {
	f := &Formatter{
		FmtScripts: make(map[string](func(io.Reader, io.Writer) error)),
	}

	// apply options
	for _, opt := range opts {
		opt(f)
	}

	return f
}

// Format runs vugufmt on input, and sends a pretty version of
// it to output. If there is an error, throw away output!
// filename is optional, but helps with generating useful output.
func (f *Formatter) Format(filename string, input io.Reader, output io.Writer) error {
	if filename == "" {
		filename = "stdin"
	}
	// First process the html bits
	doc, err := htmlx.Parse(input)
	if err != nil {
		return fmt.Errorf("failed to parse HTML5: %v", err)
	}
	if err := f.parseHTML(doc); err != nil {
		return fmt.Errorf("failed to walk HTML5: %v", err)
	}

	// At this point, re-print the parse tree.
	// I only want to print the body portion.
	var renderBody func(n *htmlx.Node) error

	renderBody = func(n *htmlx.Node) error {
		if n.Data == "body" {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if err := htmlx.Render(output, c); err != nil {
					return fmt.Errorf("failed to render HTML5: %v", err)
				}
			}
		} else if n.Data == "head" {
			return errors.New("head tag is forbidden")
		} else {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				renderBody(c)
			}
		}
		return nil
	}

	return renderBody(doc)
}

func (f *Formatter) parseHTML(n *htmlx.Node) error {
	// Clean up script blocks!
	if n.Type == htmlx.TextNode && n.Parent.Type == htmlx.ElementNode && n.Parent.Data == "script" {
		scriptType := ""
		for _, tag := range n.Parent.Attr {
			if tag.Key == "type" {
				scriptType = tag.Val
				// Handle the script type formatting.
				if scriptFmt, ok := f.FmtScripts[scriptType]; ok {
					var buf bytes.Buffer

					// Exit out on error.
					if err := scriptFmt(strings.NewReader(n.Data), &buf); err != nil {
						return err
					}

					// Save over Data with the nice version.
					n.Data = buf.String()
				}
				break
			}
		}
	}

	// Continue on with the recursive pass.
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if err := f.parseHTML(c); err != nil {
			return err
		}
	}
	return nil
}

// Diff will show differences between input and what
// Format() would do. It will return (true, nil) if there
// is a difference, (false, nil) if there is no difference,
// and (*, notnil) when the difference can't be determined.
// filename is optional, but helps with generating useful output.
func (f *Formatter) Diff(filename string, input io.Reader, output io.Writer) (bool, error) {
	if filename == "" {
		filename = "stdin"
	}

	var resBuff bytes.Buffer
	src, err := ioutil.ReadAll(input)
	if err != nil {
		return false, err
	}
	if err := f.Format(filename, bytes.NewReader(src), &resBuff); err != nil {
		return false, err
	}
	res := resBuff.Bytes()

	// No difference!
	if bytes.Equal(src, res) {
		return false, nil
	}

	// There is a difference, so what is it?
	data, err := diff(src, res, filename)
	if err != nil {
		return true, fmt.Errorf("computing diff: %s", err)
	}
	output.Write([]byte(fmt.Sprintf("diff -u %s %s\n", filepath.ToSlash(filename+".orig"), filepath.ToSlash(filename))))
	output.Write(data)
	return true, nil
}
