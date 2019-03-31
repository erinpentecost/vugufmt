package vugufmt

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

func runHTMLFmt(input io.Reader, output io.Writer, simplify bool) error {
	// First process the html bits
	doc, err := html.Parse(input)
	if err != nil {
		return fmt.Errorf("failed to parse HTML5: %v", err)
	}
	if err := parseHTML(doc, simplify); err != nil {
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

func parseHTML(n *html.Node, simplify bool) error {
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
		var buf bytes.Buffer

		// Exit out on error.
		if err := runGoFmt(strings.NewReader(n.Data), &buf, simplify); err != nil {
			return err
		}
		// Save over Data with the nice version.
		n.Data = buf.String()
	}
	// Continue on with the recursive pass.
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if err := parseHTML(c, simplify); err != nil {
			return err
		}
	}
	return nil
}
