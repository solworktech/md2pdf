[![CI][badge-build]][build]
[![GoDoc][go-docs-badge]][go-docs]
[![GoReportCard][go-report-card-badge]][go-report-card]
[![License][badge-license]][license]

## Markdown to PDF

A CLI utility which, as the name implies, generates a PDF from Markdown.

This package depends on two other packages:
- [gomarkdown](https://github.com/gomarkdown/markdown) parser to read the markdown source
- [fpdf](https://codeberg.org/go-pdf/fpdf) to generate the PDF

## Features

- [Syntax highlighting (for code blocks)](#syntax-highlighting)
- [Dark and light themes](#custom-themes)
- [Customised themes (by passing a JSON file to `md2pdf`)](#custom-themes)
- [Auto Generation of Table of Contents](#auto-generation-of-table-of-contents)
- [Support of non-Latin charsets and multiple fonts](#using-non-ascii-glyphsfonts)
- [Pagination control (using horizontal lines - especially useful for presentations)](#additional-options)
- [Page Footer (consisting of author, title and page number)](#additional-options)

## Supported Markdown elements

- Emphasised and strong text 
- Headings 1-6
- Ordered and unordered lists
- Nested lists
- Images
- Tables
- Links
- Code blocks and backticked text

## Installation 

You can obtain the pre-built `md2pdf` binary for your OS and arch
[here](https://github.com/solworktech/md2pdf/releases); 
you can also install the `md2pdf` binary directly onto your `$GOBIN` dir with:

```sh
$ go install github.com/solworktech/md2pdf/v2/cmd/md2pdf@latest
```

`md2pdf` is also available via [Homebrew](https://formulae.brew.sh/formula/md2pdf):

```sh
$ brew install md2pdf
```

## Syntax highlighting

`md2pdf` supports colourised output via the [gohighlight module](https://github.com/jessp01/gohighlight).

For examples, see [testdata/syntax_highlighting.md](./testdata/syntax_highlighting.md) and 
[testdata/syntax_highlighting.pdf](./testdata/syntax_highlighting.pdf)

## Custom themes

`md2pdf` supports both light and dark themes out of the box (use `--theme light` or `--theme dark` - no config required). 

However, if you wish to customise the font faces, sizes and colours, you can use the JSONs in
[custom_themes](./custom_themes) as a starting point. Edit to your liking and pass `--theme /path/to/json` to `md2pdf`

## Auto Generation of Table of Contents

`md2pdf` can automatically generate a TOC where each item corresponds to a header in the doc and include it in the first page.
TOC items can then be clicked to navigate to the relevant section (similar to HTML `<a>` anchors).

To make use of this feature, simply pass `--generate-toc` as an argument.

## Quick start

```
$ cd cmd/md2pdf
$ go run md2pdf.go -i test.md -o test.pdf
```

To benefit from Syntax highlighting, invoke thusly:

```
$ go run md2pdf.go -i syn_test.md -s /path/to/syntax_files -o test.pdf
```

To convert multiple MD files into a single PDF, use:
```
$ go run md2pdf.go -i /path/to/md/directory -o test.pdf
```

This repo has the [gohighlight module](https://github.com/jessp01/gohighlight) configured as a submodule, so if you clone
with `--recursive`, you will have the `highlight` dir in its root. Alternatively, you may issue the following command to update an
existing clone:

```sh
git submodule update --remote  --init
```

*Note 1: the `cmd` folder has an example for the syntax highlighting. 
See the script `run_syntax_highlighting.sh`. This example assumes that
the folder with the syntax files is located at a relative location:
`../../../jessp01/gohighlight/syntax_files`.*

*Note 2: when annotating the code block to specify the language, the
annotation name must match the syntax base filename.*

### Additional options

```sh
  -author string
    	Author name; used if -footer is passed
  -font-file string
    	path to font file to use
  -font-name string
    	Font name ID; e.g 'Helvetica-1251'
  -generate-toc
    	Auto Generate Table of Contents (TOC)
  -help
    	Show usage message
  -i string
    	Input filename, dir consisting of .md|.markdown files or HTTP(s) URL; default is os.Stdin
  -log-file string
    	Path to log file
  -new-page-on-hr
    	Interpret HR as a new page; useful for presentations
  -o string
    	Output PDF filename; required
  -orientation string
    	[portrait | landscape] (default "portrait")
  -page-size string
    	[A3 | A4 | A5] (default "A4")
  -s string
    	Path to github.com/jessp01/gohighlight/syntax_files
  -theme string
    	[light | dark | /path/to/custom/theme.json] (default "light")
  -title string
    	Presentation title
  -unicode-encoding string
    	e.g 'cp1251'
  -version
    	Print version and build info
  -with-footer
    	Print doc footer (<author>  <title>  <page number>)
```

For example, the below will:

- Set the title to `My Grand Title`
- Set `Random Bloke` as the author (used in the footer)
- Set the dark theme
- Start a new page when encountering an HR (`---`); useful for creating presentations
- Print a footer (`author name, title, page number`)

```sh
$ go run md2pdf.go  -i /path/to/md \
    -o /path/to/pdf --title "My Grand Title" --author "Random Bloke" \
    --theme dark --new-page-on-hr --with-footer
```

## Using non-ASCII Glyphs/Fonts

To use a non-ASCII language, the PDF generator must be configured with `WithUnicodeTranslator`:

```go
// https://en.wikipedia.org/wiki/Windows-1251
pf := mdtopdf.NewPdfRenderer("", "", *output, "trace.log", mdtopdf.WithUnicodeTranslator("cp1251")) 
```

In addition, this package's `Styler` must be used to set the font to match what is configured with the PDF generator.

A complete working example can be found for Russian in the `cmd` folder named
`russian.go`.

For a full example, run:

```sh
$ go run md2pdf.go -i russian.md -o russian.pdf \
    --unicode-encoding cp1251 --font-file helvetica_1251.json --font-name Helvetica_1251
```

## Tests

The tests included in this repo (see the `testdata` folder) were taken from the BlackFriday package.
While the tests may complete without errors, visual inspection of the created PDF is the
only way to determine if the tests *really* pass!

The tests create log files that trace the [gomarkdown](https://github.com/gomarkdown/markdown) parser
callbacks. This is a valuable debugging tool, showing each callback 
and the data provided while the AST is presented.

## Limitations and Known Issues

- It is common for Markdown to include HTML. HTML is treated as a "code block". *There is no attempt to convert raw HTML to PDF.*
- Github-flavoured Markdown permits strikethrough using tildes. This is not supported by `fpdf` as a font style at present.
- The markdown link title (which would show when converted to HTML as hover-over text) is not supported. The generated PDF will show the URL, but this is a function of the PDF viewer.
- Definition lists are not supported
- The following text features may be tweaked: font, size, spacing, style, fill colour, and text colour. These are exported and available via the `Styler` struct. Note that fill colour only works when using `CellFormat()`. This is the case for tables, code blocks, and backticked text.

## Contributions

- Set up and run pre-commit hooks:

```sh
# Install the needed GO packages:
go install github.com/go-critic/go-critic/cmd/gocritic@latest
go install golang.org/x/tools/cmd/goimports@latest
go install golang.org/x/lint/golint@latest
go install github.com/gordonklaus/ineffassign@latest

# Install the `pre-commit` util:
pip install pre-commit

# Generate `.git/hooks/pre-commit`:
pre-commit install
```

Following that, these tests will run every time you invoke `git commit`:
```sh
go fmt...................................................................Passed
go imports...............................................................Passed
go vet...................................................................Passed
go lint..................................................................Passed
go-critic................................................................Passed
```

- Submit a pull request and include a succinct description of the feature or issue it addresses 

[license]: ./LICENSE
[badge-license]: https://img.shields.io/github/license/solworktech/md2pdf.svg
[go-docs-badge]: https://godoc.org/github.com/solworktech/md2pdf?status.svg
[go-docs]: https://godoc.org/github.com/solworktech/md2pdf/v2
[badge-build]: https://github.com/solworktech/md2pdf/actions/workflows/go.yml/badge.svg
[build]: https://github.com/solworktech/md2pdf/actions/workflows/go.yml
[go-report-card-badge]: https://goreportcard.com/badge/github.com/solworktech/md2pdf/v2
[go-report-card]: https://goreportcard.com/report/github.com/solworktech/md2pdf/v2
