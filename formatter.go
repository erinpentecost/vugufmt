package vugufmt

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

// Formatter allows you to format vugu files.
type Formatter struct {
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
func (f *Formatter) Format(input io.Reader, output io.Writer) error {
	// First process the html bits
	doc, err := html.Parse(input)
	if err != nil {
		return fmt.Errorf("failed to parse HTML5: %v", err)
	}
	if err := f.parseHTML(doc); err != nil {
		return fmt.Errorf("failed to walk HTML5: %v", err)
	}

	// At this point, re-print the parse tree.
	// I only want to print the body portion.
	var renderBody func(n *html.Node) error

	renderBody = func(n *html.Node) error {
		if n.Data == "body" {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if err := html.Render(output, c); err != nil {
					return fmt.Errorf("failed to render HTML5: %v", err)
				}
			}
		} else {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				renderBody(c)
			}
		}
		return nil
	}

	return renderBody(doc)
}

func (f *Formatter) parseHTML(n *html.Node) error {
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
	if scriptFmt, ok := f.FmtScripts[scriptType]; ok {
		var buf bytes.Buffer

		//fmt.Printf("\n\n\t%s:\n\n%s\n\n", scriptType, n.Data)

		// Exit out on error.
		if err := scriptFmt(strings.NewReader(n.Data), &buf); err != nil {
			return err
		}

		// Save over Data with the nice version.
		n.Data = buf.String()
	}

	// Continue on with the recursive pass.
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if err := f.parseHTML(c); err != nil {
			return err
		}
	}
	return nil
}
