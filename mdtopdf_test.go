/*
 * Markdown to PDF Converter
 * Available at http://github.com/mandolyte/mdtopdf
 *
 * Copyright © 2018 Cecil New <cecil.new@gmail.com>.
 * Distributed under the MIT License.
 * See README.md for details.
 *
 * Dependencies
 * This package depends on two other packages:
 *
 * Go Markdown processor
 *   Available at https://github.com/gomarkdown/markdown
 *
 * fpdf - a PDF document generator with high level support for
 *   text, drawing and images.
 *   Available at https://github.com/go-pdf/fpdf
 */

package mdtopdf

import (
	"os"
	"path"
	"strings"
	"testing"
)

func testit(inputf string, gohighlight bool, t *testing.T) {
	inputd := "./testdata/"
	input := path.Join(inputd, inputf)

	tracerfile := path.Join(inputd, strings.TrimSuffix(path.Base(input), ".text"))
	tracerfile += ".log"

	pdffile := path.Join(inputd, strings.TrimSuffix(path.Base(input), ".text"))
	pdffile += ".pdf"

	content, err := os.ReadFile(input)
	if err != nil {
		t.Errorf("%v:%v", input, err)
	}

	var r *PdfRenderer
	var opts []RenderOption
	if gohighlight {
		opts = []RenderOption{IsHorizontalRuleNewPage(true), SetSyntaxHighlightBaseDir("./highlight/syntax_files")}
	}
	params := PdfRendererParams{
		Orientation: "",
		Papersz:     "",
		PdfFile:     pdffile,
		TracerFile:  tracerfile,
		Opts:        opts,
		Theme:       LIGHT,
	}
	r = NewPdfRenderer(params)
	err = r.Process(content)
	if err != nil {
		t.Error(err)
	}
}

func TestTables(t *testing.T) {
	testit("Tables.text", false, t)
}

func TestMarkdownDocumenationBasic(t *testing.T) {
	testit("Markdown Documentation - Basics.text", false, t)
}

func TestMarkdownDocumenationSyntax(t *testing.T) {
	testit("Markdown Documentation - Syntax.text", false, t)
}

func TestMarkdownDocumenationColourSyntax(t *testing.T) {
	testit("Markdown Documentation - Colour.text", true, t)
}

func TestImage(t *testing.T) {
	testit("Image.text", false, t)
}

func TestAutoLinks(t *testing.T) {
	testit("Auto links.text", false, t)
}

func TestAmpersandEncoding(t *testing.T) {
	testit("Amps and angle encoding.text", false, t)
}

func TestInlineLinks(t *testing.T) {
	testit("Links, inline style.text", false, t)
}

func TestLists(t *testing.T) {
	testit("Ordered and unordered lists.text", false, t)
}

func TestStringEmph(t *testing.T) {
	testit("Strong and em together.text", false, t)
}

func TestTabs(t *testing.T) {
	testit("Tabs.text", false, t)
}

func TestBackslashEscapes(t *testing.T) {
	testit("Backslash escapes.text", false, t)
}

func TestBackquotes(t *testing.T) {
	testit("Blockquotes with code blocks.text", false, t)
}

func TestCodeBlocks(t *testing.T) {
	testit("Code Blocks.text", false, t)
}

func TestCodeSpans(t *testing.T) {
	testit("Code Spans.text", false, t)
}

func TestHardWrappedPara(t *testing.T) {
	testit("Hard-wrapped paragraphs with list-like lines no empty line before block.text", false, t)
}

func TestHardWrappedPara2(t *testing.T) {
	testit("Hard-wrapped paragraphs with list-like lines.text", false, t)
}

func TestHorizontalRules(t *testing.T) {
	testit("Horizontal rules.text", false, t)
}

func TestInlineHtmlSimple(t *testing.T) {
	testit("Inline HTML (Simple).text", false, t)
}

func TestInlineHtmlAdvanced(t *testing.T) {
	testit("Inline HTML (Advanced).text", false, t)
}

func TestInlineHtmlComments(t *testing.T) {
	testit("Inline HTML comments.text", false, t)
}

func TestTitleWithQuotes(t *testing.T) {
	testit("Literal quotes in titles.text", false, t)
}

func TestNestedBlockquotes(t *testing.T) {
	testit("Nested blockquotes.text", false, t)
}

func TestLinksReference(t *testing.T) {
	testit("Links, reference style.text", false, t)
}

func TestLinksShortcut(t *testing.T) {
	testit("Links, shortcut references.text", false, t)
}

func TestTidyness(t *testing.T) {
	testit("Tidyness.text", false, t)
}
