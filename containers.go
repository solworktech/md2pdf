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

package mdtopdf

type listType int

const (
	notlist listType = iota
	unordered
	ordered
	definition
)

// This slice of float64 contains the width of each cell
// in the header of a table. These will be the widths used
// in the table body as well.
var cellwidths []float64
var curdatacell int
var fill = false
var incell = false

func (n listType) String() string {
	switch n {
	case notlist:
		return "Not a List"
	case unordered:
		return "Unordered"
	case ordered:
		return "Ordered"
	case definition:
		return "Definition"
	}
	return ""
}

type containerState struct {
	textStyle      Styler
	leftMargin     float64
	firstParagraph bool

	// populated if node type is a list
	listkind   listType
	itemNumber int // only if an ordered list

	// populated if node type is a link
	destination string

	// populated if table cell
	isHeader bool

	// populated if table cell (apply styles first)
	cellInnerString      string
	cellInnerStringStyle *Styler
}

type states struct {
	stack []*containerState
}

func (s *states) push(c *containerState) {
	s.stack = append(s.stack, c)
}

func (s *states) pop() *containerState {
	var x *containerState
	x, s.stack = s.stack[len(s.stack)-1], s.stack[:len(s.stack)-1]
	return x
}

func (s *states) peek() *containerState {
	return s.stack[len(s.stack)-1]
}

func (s *states) parent() *containerState {
	return s.stack[len(s.stack)-2]
}
