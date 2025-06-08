/*
 * Markdown to PDF Converter
 * Available at http://github.com/solworktech/md2pdf
 *
 * Copyright © Cecil New <cecil.new@gmail.com>, Jesse Portnoy <jesse@packman.io>.
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
 *   Available at https://codeberg.org/go-pdf/fpdf
 */

// Package mdtopdf converts markdown to PDF.
package mdtopdf

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"strings"

	"codeberg.org/go-pdf/fpdf"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

// Color is a RGB set of ints; for a nice picker
// see https://www.w3schools.com/colors/colors_picker.asp
type Color struct {
	Red, Green, Blue int
}

// Styler is the struct to capture the styling features for text
// Size and Spacing are specified in points.
// The sum of Size and Spacing is used as line height value
// in the fpdf API
type Styler struct {
	Font      string
	Style     string
	Size      float64
	Spacing   float64
	TextColor Color
	FillColor Color
}

// RenderOption allows to define functions to configure the renderer
type RenderOption func(r *PdfRenderer)

// Theme [light|dark]
type Theme int

const (
	// DARK theme const
	DARK Theme = 1
	// LIGHT theme const
	LIGHT Theme = 2
	// CUSTOM theme const
	CUSTOM Theme = 3
)

// PdfRenderer is the struct to manage conversion of a markdown object
// to PDF format.
type PdfRenderer struct {
	// Pdf can be used to access the underlying created fpdf object
	// prior to processing the markdown source
	Pdf                *fpdf.Fpdf
	orientation, units string
	papersize, fontdir string

	// trace/log file if present
	pdfFile, tracerFile string
	w                   *bufio.Writer

	// default margins for safe keeping
	mleft, mtop, mright, mbottom float64

	// normal text
	Normal            Styler
	em                float64
	unicodeTranslator func(string) string

	// link text
	Link Styler

	// backticked text
	Backtick Styler

	// blockquote text
	Blockquote  Styler
	IndentValue float64

	// Headings
	H1 Styler
	H2 Styler
	H3 Styler
	H4 Styler
	H5 Styler
	H6 Styler

	// Table styling
	THeader Styler
	TBody   Styler

	cs states

	// code styling
	Code Styler

	// update styling
	NeedCodeStyleUpdate       bool
	NeedBlockquoteStyleUpdate bool
	HorizontalRuleNewPage     bool
	SyntaxHighlightBaseDir    string
	InputBaseURL              string
	Theme                     Theme
	BackgroundColor           Color
	documentMatter            ast.DocumentMatters // keep track of front/main/back matter.
	Extensions                parser.Extensions
	ColumnWidths              map[ast.Node][]float64
}

// TOCEntry represents a table of contents entry
type TOCEntry struct {
	Level int
	Title string
	ID    string
}

// TOCVisitor implements ast.NodeVisitor to collect headers
type TOCVisitor struct {
	entries []TOCEntry
}

// Visit implements the ast.NodeVisitor interface
func (v *TOCVisitor) Visit(node ast.Node, entering bool) ast.WalkStatus {
	if !entering {
		return ast.GoToNext
	}

	// Check if the node is a heading
	if heading, ok := node.(*ast.Heading); ok {
		// Extract the text content from the heading
		title := ExtractTextFromNode(heading)
		if title != "" {
			// Create a simple ID from the title (lowercase, replace spaces with hyphens)
			id := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(title), " ", "-"))
			// Remove special characters for cleaner IDs
			id = strings.ReplaceAll(id, ".", "")
			id = strings.ReplaceAll(id, ",", "")
			id = strings.ReplaceAll(id, "!", "")
			id = strings.ReplaceAll(id, "?", "")

			entry := TOCEntry{
				Level: heading.Level,
				Title: title,
				ID:    id,
			}
			v.entries = append(v.entries, entry)
		}
	}

	return ast.GoToNext
}

// ExtractTextFromNode recursively extracts text content from AST nodes
func ExtractTextFromNode(node ast.Node) string {
	var text strings.Builder

	ast.WalkFunc(node, func(node ast.Node, entering bool) ast.WalkStatus {
		if entering {
			switch n := node.(type) {
			case *ast.Text:
				text.Write(n.Literal)
			case *ast.Code:
				text.Write(n.Literal)
			}
		}
		return ast.GoToNext
	})

	return text.String()
}

// GetTOCEntries returns TOC entries
func GetTOCEntries(content []byte) ([]TOCEntry, error) {

	// Create parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)

	// Parse the markdown content
	doc := markdown.Parse(content, p)

	// Create visitor to collect TOC entries
	visitor := &TOCVisitor{}

	// Walk the AST and collect headers
	ast.Walk(doc, visitor)

	return visitor.entries, nil
}

// SetLightTheme sets theme to 'light'
func (r *PdfRenderer) SetLightTheme() {
	r.BackgroundColor = Colorlookup("white")
	r.SetPageBackground("", r.BackgroundColor)
	// Normal Text
	r.Normal = Styler{Font: "Arial", Style: "", Size: 12, Spacing: 2,
		TextColor: Colorlookup("black"), FillColor: Colorlookup("white")}

	// Link text
	r.Link = Styler{Font: "Arial", Style: "b", Size: 12, Spacing: 2,
		TextColor: Colorlookup("cornflowerblue")}

	// Backticked text
	r.Backtick = Styler{Font: "Times", Style: "", Size: 12, Spacing: 2,
		TextColor: Color{37, 27, 14}, FillColor: Color{200, 200, 200}}

	// Quoted Text

	r.Blockquote = Styler{Font: "Times", Style: "", Size: 12, Spacing: 2,
		TextColor: Color{37, 27, 14}, FillColor: Color{200, 200, 200}}

	// Code text
	r.Code = Styler{Font: "Times", Style: "", Size: 12, Spacing: 2,
		TextColor: Color{37, 27, 14}, FillColor: Color{200, 200, 200}}

	// Headings
	r.H1 = Styler{Font: "Arial", Style: "b", Size: 24, Spacing: 5,
		TextColor: Colorlookup("black"), FillColor: Colorlookup("white")}
	r.H2 = Styler{Font: "Arial", Style: "b", Size: 22, Spacing: 5,
		TextColor: Colorlookup("black"), FillColor: Colorlookup("white")}
	r.H3 = Styler{Font: "Arial", Style: "b", Size: 20, Spacing: 5,
		TextColor: Colorlookup("black"), FillColor: Colorlookup("white")}
	r.H4 = Styler{Font: "Arial", Style: "b", Size: 18, Spacing: 5,
		TextColor: Colorlookup("black"), FillColor: Colorlookup("white")}
	r.H5 = Styler{Font: "Arial", Style: "b", Size: 16, Spacing: 5,
		TextColor: Colorlookup("black"), FillColor: Colorlookup("white")}
	r.H6 = Styler{Font: "Arial", Style: "b", Size: 14, Spacing: 5,
		TextColor: Colorlookup("black"), FillColor: Colorlookup("white")}

	r.Blockquote = Styler{Font: "Arial", Style: "i", Size: 12, Spacing: 2,
		TextColor: Colorlookup("black"), FillColor: Colorlookup("white")}

	// Table Header Text
	r.THeader = Styler{Font: "Arial", Style: "b", Size: 12, Spacing: 2,
		TextColor: Colorlookup("black"), FillColor: Color{180, 180, 180}}

	// Table Body Text
	r.TBody = Styler{Font: "Arial", Style: "", Size: 12, Spacing: 2,
		TextColor: Colorlookup("black"), FillColor: Color{240, 240, 240}}
}

// SetDarkTheme sets theme to 'dark'
func (r *PdfRenderer) SetDarkTheme() {
	r.BackgroundColor = Colorlookup("black")
	r.SetPageBackground("", r.BackgroundColor)
	// Normal Text
	r.Normal = Styler{Font: "Arial", Style: "", Size: 12, Spacing: 2,
		FillColor: Colorlookup("black"), TextColor: Colorlookup("white")}

	// Quoted Text
	r.Blockquote = Styler{Font: "Arial", Style: "", Size: 12, Spacing: 2,
		FillColor: Colorlookup("black"), TextColor: Colorlookup("white")}

	// Link text
	r.Link = Styler{Font: "Arial", Style: "b", Size: 12, Spacing: 2,
		TextColor: Colorlookup("cornflowerblue")}

	// Backticked text
	r.Backtick = Styler{Font: "Times", Style: "", Size: 12, Spacing: 2,
		TextColor: Colorlookup("lightgrey"), FillColor: Color{32, 35, 37}}

	// Code text
	r.Code = Styler{Font: "Times", Style: "", Size: 12, Spacing: 2,
		TextColor: Colorlookup("lightgrey"), FillColor: Color{32, 35, 37}}

	// Headings
	r.H1 = Styler{Font: "Arial", Style: "b", Size: 24, Spacing: 5,
		FillColor: Colorlookup("black"), TextColor: Colorlookup("darkgray")}
	r.H2 = Styler{Font: "Arial", Style: "b", Size: 22, Spacing: 5,
		FillColor: Colorlookup("black"), TextColor: Colorlookup("darkgray")}
	r.H3 = Styler{Font: "Arial", Style: "b", Size: 20, Spacing: 5,
		FillColor: Colorlookup("black"), TextColor: Colorlookup("darkgray")}
	r.H4 = Styler{Font: "Arial", Style: "b", Size: 18, Spacing: 5,
		FillColor: Colorlookup("black"), TextColor: Colorlookup("darkgray")}
	r.H5 = Styler{Font: "Arial", Style: "b", Size: 16, Spacing: 5,
		FillColor: Colorlookup("black"), TextColor: Colorlookup("darkgray")}
	r.H6 = Styler{Font: "Arial", Style: "b", Size: 14, Spacing: 5,
		FillColor: Colorlookup("black"), TextColor: Colorlookup("darkgray")}

	r.Blockquote = Styler{Font: "Arial", Style: "i", Size: 12, Spacing: 2,
		FillColor: Colorlookup("black"), TextColor: Colorlookup("darkgray")}

	// Table Header Text
	r.THeader = Styler{Font: "Arial", Style: "b", Size: 12, Spacing: 2,
		TextColor: Colorlookup("darkgray"), FillColor: Color{27, 27, 27}}

	// Table Body Text
	r.TBody = Styler{Font: "Arial", Style: "", Size: 12, Spacing: 2,
		FillColor: Color{200, 200, 200}, TextColor: Color{128, 128, 128}}

}

// SetCustomTheme sets a custom theme based on JSON config
func (r *PdfRenderer) SetCustomTheme(themeJSONFile string) {

	config, err := os.ReadFile(themeJSONFile)
	if err != nil {
		log.Fatal(err)
	}
	// Fill the instance from the JSON file content
	err = json.Unmarshal(config, &r)
	// Check if is there any error while filling the instance
	if err != nil {
		log.Fatal("Error parsing ", themeJSONFile, ":\n", err)
	}
}

// PdfRendererParams struct to hold params passed to NewPdfRenderer
type PdfRendererParams struct {
	Orientation, Papersz, PdfFile, TracerFile, FontFile, FontName string
	Opts                                                          []RenderOption
	Theme                                                         Theme
	CustomThemeFile                                               string
}

// NewPdfRenderer creates and configures an PdfRenderer object,
// which satisfies the Renderer interface.
func NewPdfRenderer(params PdfRendererParams) *PdfRenderer {

	r := new(PdfRenderer)

	// set filenames
	r.pdfFile = params.PdfFile
	r.tracerFile = params.TracerFile

	// Global things
	r.orientation = "portrait"
	if params.Orientation != "" {
		r.orientation = params.Orientation
	}

	r.units = "pt"
	r.papersize = "Letter"
	if params.Papersz != "" {
		r.papersize = params.Papersz
	}

	r.fontdir = "."

	r.Theme = params.Theme

	r.Pdf = fpdf.New(r.orientation, r.units, r.papersize, r.fontdir)

	r.Pdf.SetHeaderFunc(func() {
		r.SetPageBackground("", r.BackgroundColor)
	})

	switch r.Theme {
	case DARK:
		r.SetDarkTheme()
	case LIGHT:
		r.SetLightTheme()
	case CUSTOM:
		if params.CustomThemeFile != "" {
			r.SetCustomTheme(params.CustomThemeFile)
		}
	}
	r.Pdf.AddPage()
	// set default font
	r.setStyler(r.Normal)
	r.mleft, r.mtop, r.mright, r.mbottom = r.Pdf.GetMargins()
	r.em = r.Pdf.GetStringWidth("m")
	r.IndentValue = 3 * r.em

	r.cs = states{stack: make([]*containerState, 0)}
	initcurrent := &containerState{
		listkind:  notlist,
		textStyle: r.Normal, leftMargin: r.mleft}
	r.cs.push(initcurrent)

	for _, o := range params.Opts {
		o(r)
	}

	return r
}

// NewPdfRendererWithDefaultStyler creates and configures an PdfRenderer object,
// which satisfies the Renderer interface.
// update default styler for normal
func NewPdfRendererWithDefaultStyler(orient, papersz, pdfFile, tracerFile string, defaultStyler Styler, opts []RenderOption, theme Theme) *PdfRenderer {
	opts = append(opts, func(r *PdfRenderer) {
		r.Normal = defaultStyler
	})
	params := PdfRendererParams{
		Orientation: orient,
		Papersz:     papersz,
		PdfFile:     pdfFile,
		TracerFile:  tracerFile,
		Opts:        opts,
		Theme:       theme,
	}

	return NewPdfRenderer(params)
}

// Process takes the markdown content, parses it to generate the PDF
func (r *PdfRenderer) Process(content []byte) error {
	// try to open tracer
	var f *os.File
	var err error
	if r.tracerFile != "" {
		f, err = os.Create(r.tracerFile)
		if err != nil {
			return fmt.Errorf("os.Create() on tracefile error:%v", err)
		}
		defer f.Close()
		r.w = bufio.NewWriter(f)
		defer r.w.Flush()
	}

	err = r.Run(content)
	if err != nil {
		return fmt.Errorf("error on %v:%v", r.pdfFile, err)
	}

	err = r.Pdf.OutputFileAndClose(r.pdfFile)
	if err != nil {
		return fmt.Errorf("error on %v:%v", r.pdfFile, err)
	}

	return nil
}

// Run takes the markdown content, parses it but don't generate the PDF. you can access the PDF with youRenderer.Pdf
func (r *PdfRenderer) Run(content []byte) error {
	// Preprocess content by changing all CRLF to LF
	s := content
	s = markdown.NormalizeNewlines(s)

	if r.unicodeTranslator != nil {
		s = []byte(r.unicodeTranslator(string(s)))
	}

	p := parser.NewWithExtensions(r.Extensions)
	doc := markdown.Parse(s, p)

	setColumnWidths(doc, r)
	_ = markdown.Render(doc, r)

	return nil
}

// Parses all tables and sets the column width to the longest string in that column
func setColumnWidths(doc ast.Node, r *PdfRenderer) {
	columnWidths := map[ast.Node][]float64{}
	intable := false
	inheader := true
	cellnum := 0
	lengths := []float64{}
	textlength := float64(0)
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		switch n := node.(type) {
		case *ast.Table:
			if entering {
				intable = true
			} else {
				intable = false
				columnWidths[node] = lengths
			}

		case *ast.TableHeader:
			inheader = entering
			if entering {
				lengths = []float64{}
			}
		case *ast.TableRow:
			if entering {
				cellnum = 0
			}
		case *ast.TableCell:
			if entering {
				if inheader {
					lengths = append(lengths, 0)
				}
			} else {
				textlength += textlength * 0.2

				currentMax := lengths[cellnum]
				if textlength > currentMax {
					lengths[cellnum] = textlength
				}
				textlength = 0
				cellnum++
			}
		case *ast.Text:
			if entering && intable {
				l := r.Pdf.GetStringWidth(string(n.Literal))
				textlength += l
			}
		}
		return ast.GoToNext
	})
	r.ColumnWidths = columnWidths
}

// UpdateParagraphStyler - update with default styler
func (r *PdfRenderer) UpdateParagraphStyler(defaultStyler Styler) {
	initcurrent := &containerState{
		listkind:  notlist,
		textStyle: defaultStyler, leftMargin: r.mleft}
	r.cs.push(initcurrent)
}

// UpdateCodeStyler - update code fill styler
func (r *PdfRenderer) UpdateCodeStyler() {
	r.NeedCodeStyleUpdate = true
}

// UpdateBlockquoteStyler - update Blockquote fill styler
func (r *PdfRenderer) UpdateBlockquoteStyler() {
	r.NeedBlockquoteStyleUpdate = true
}

func (r *PdfRenderer) setStyler(s Styler) {
	// see https://github.com/solworktech/md2pdf/issues/18#issuecomment-2179694815
	// This does not address the root cause
	// (https://github.com/solworktech/md2pdf/issues/18#issuecomment-2179694815)
	// but it will correct all cases and is safer.
	if s.Style == "bb" {
		s.Style = "b"
	}
	r.Pdf.SetFont(s.Font, s.Style, s.Size)
	r.Pdf.SetTextColor(s.TextColor.Red, s.TextColor.Green, s.TextColor.Blue)
	r.Pdf.SetFillColor(s.FillColor.Red, s.FillColor.Green, s.FillColor.Blue)
}

func (r *PdfRenderer) write(s Styler, t string) {
	// fmt.Printf("%s, %#v\n",t, s)
	r.Pdf.Write(s.Size+s.Spacing, t)
}

func (r *PdfRenderer) multiCell(s Styler, t string) {
	r.Pdf.MultiCell(0, s.Size+s.Spacing, t, "", "", true)
}

func (r *PdfRenderer) writeLink(s Styler, display, url string) {
	r.Pdf.WriteLinkString(s.Size+s.Spacing, display, url)
}

// RenderNode is a default renderer of a single node of a syntax tree. For
// block nodes it will be called twice: first time with entering=true, second
// time with entering=false, so that it could know when it's working on an open
// tag and when on close. It writes the result to w.
//
// The return value is a way to tell the calling walker to adjust its walk
// pattern: e.g. it can terminate the traversal by returning Terminate. Or it
// can ask the walker to skip a subtree of this node by returning SkipChildren.
// The typical behavior is to return GoToNext, which asks for the usual
// traversal to the next node.
// (above taken verbatim from the blackfriday v2 package)
func (r *PdfRenderer) RenderNode(w io.Writer, node ast.Node, entering bool) ast.WalkStatus {
	switch node := node.(type) {
	case *ast.Text:
		r.processText(node)
	case *ast.Softbreak:
		r.tracer("Softbreak", "Output newline")
		r.cr()
	case *ast.Hardbreak:
		r.tracer("Hardbreak", "Output newline")
		r.cr()
	case *ast.Emph:
		r.processEmph(node, entering)
	case *ast.Strong:
		r.processStrong(node, entering)
	case *ast.Del:
		if entering {
			r.tracer("DEL (entering)", "Not handled")
		} else {
			r.tracer("DEL (leaving)", "Not handled")
		}
	case *ast.HTMLSpan:
		r.tracer("HTMLSpan", "Not handled")
	case *ast.Link:
		r.processLink(*node, entering)
	case *ast.Image:
		r.processImage(*node, entering)
	case *ast.Code:
		r.processCode(node)
	case *ast.Document:
		r.tracer("Document", "Not Handled")
	case *ast.Paragraph:
		r.processParagraph(node, entering)
	case *ast.BlockQuote:
		r.processBlockQuote(node, entering)
	case *ast.HTMLBlock:
		r.processHTMLBlock(node)
	case *ast.Heading:
		r.processHeading(*node, entering)
	case *ast.HorizontalRule:
		r.processHorizontalRule(node)
	case *ast.List:
		r.processList(*node, entering)
	case *ast.ListItem:
		r.processItem(*node, entering)
	case *ast.CodeBlock:
		r.processCodeblock(*node)
	case *ast.Table:
		r.processTable(node, entering)
	case *ast.TableHeader:
		r.processTableHead(node, entering)
	case *ast.TableBody:
		r.processTableBody(node, entering)
	case *ast.TableRow:
		r.processTableRow(node, entering)
	case *ast.TableCell:
		r.processTableCell(*node, entering)
	/*case *ast.Math:
	r.processMath(node)*/
	default:
		fmt.Printf("Unknown node type: %T. Skipping\n", node)
	}
	return ast.GoToNext
}

// RenderHeader is not supported.
func (r *PdfRenderer) RenderHeader(w io.Writer, ast ast.Node) {
	r.tracer("RenderHeader", "Not handled")
}

// RenderFooter is not supported.
func (r *PdfRenderer) RenderFooter(w io.Writer, _ ast.Node) {
}

func (r *PdfRenderer) cr() {
	LH := r.cs.peek().textStyle.Size + r.cs.peek().textStyle.Spacing
	r.tracer("cr()", fmt.Sprintf("LH=%v", LH))
	r.write(r.cs.peek().textStyle, "\n")
}

// Tracer traces parse and pdf generation activity.
func (r *PdfRenderer) tracer(source, msg string) {
	if r.tracerFile != "" {
		indent := strings.Repeat("-", len(r.cs.stack)-1)
		r.w.WriteString(fmt.Sprintf("%v[%v] %v\n", indent, source, msg))
	}
}

func dorect(doc *fpdf.Fpdf, x, y, w, h float64, color Color) {
	doc.SetFillColor(color.Red, color.Green, color.Blue)
	doc.Rect(x, y, w, h, "F")
}

// SetPageBackground - sets background colour of page. String IDs ("blue", "grey", etc) and `Color` structs are both supported
func (r *PdfRenderer) SetPageBackground(colorStr string, color Color) {
	w, h := r.Pdf.GetPageSize()
	if colorStr != "" {
		color = Colorlookup(colorStr)
	}
	dorect(r.Pdf, 0, 0, w, h, color)
}

// Options

// WithUnicodeTranslator configures a unico translator to support characters for latin, russian, etc..
func WithUnicodeTranslator(cp string) RenderOption {
	return func(r *PdfRenderer) {
		r.unicodeTranslator = r.Pdf.UnicodeTranslatorFromDescriptor(cp)
	}
}

// IsHorizontalRuleNewPage if true, will start a new page when encountering a HR (---). Useful for presentations.
func IsHorizontalRuleNewPage(value bool) RenderOption {
	return func(r *PdfRenderer) {
		r.HorizontalRuleNewPage = value
	}
}

// SetSyntaxHighlightBaseDir path to https://github.com/jessp01/gohighlight/tree/master/syntax_files
func SetSyntaxHighlightBaseDir(path string) RenderOption {
	return func(r *PdfRenderer) {
		r.SyntaxHighlightBaseDir = path
	}
}
