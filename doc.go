/*
Package mdtopdf implements a PDF document generator for markdown documents.

# Introduction

This package depends on two other packages:

* The [gomarkdown](https://github.com/gomarkdown/markdown) parser to read the markdown source

* The fpdf package to generate the PDF

The tests included here are from the BlackFriday package.
See the "testdata" folder.
The tests create PDF files and thus while the tests may complete
without errors, visual inspection of the created PDF is the
only way to determine if the tests *really* pass!

The tests create log files that trace the BlackFriday parser
callbacks. This is a valuable debug tool showing each callback
and data provided in each while the AST is presented.

# Installation

To install the package:

	go get github.com/solworktech/md2pdf/v2

# Quick start

In the cmd folder is an example using the package. It demonstrates
a number of features. The test PDF was created with this command:

	go run convert.go -i test.md -o test.pdf

See README for limitations and known issues
*/
package mdtopdf
