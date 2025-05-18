
[![CI][badge-build]][build]
[![GoDoc][go-docs-badge]][go-docs]
[![GoReportCard][go-report-card-badge]][go-report-card]
[![License][badge-license]][license]

## Markdown to PDF

A CLI utility which, as the name implies, generates PDF from Markdown.

This package depends on two other packages:
- [gomarkdown](https://github.com/gomarkdown/markdown) parser to read the markdown source
- [fpdf](https://github.com/go-pdf/fpdf) to generate the PDF

## Features

- [Syntax highlighting (for code blocks)](#syntax-highlighting)
- [Dark and light themes](#custom-themes)
- [Customised themes (by passing a JSON file to `md2pdf`)](#custom-themes)
- [Support of non-Latin charsets and multiple fonts](#using-non-ascii-glyphsfonts)
- [Pagination control (using horizontal lines - especially useful for presentations)](#additional-options)
- [Page Footer (consisting of author, title and page number)](#additional-options)

## Supported Markdown elements

- Emphasized and strong text 
- Headings 1-6
- Ordered and unordered lists
- Nested lists
- Images
- Tables
- Links
- Code blocks and backticked text

## Installation 

You can obtain the pre-built `md2pdf` binary for your OS and arch
[here](https://github.com/solworktech/mdtopdf/releases); 
you can also install the `md2pdf` binary directly onto your `$GOBIN` dir with:

```sh
$ go install github.com/mandolyte/mdtopdf/v2/cmd/md2pdf@latest
```

`md2pdf` is also available via [Homebrew](https://formulae.brew.sh/formula/md2pdf):

```sh
$ brew install md2pdf
```

## Syntax highlighting

`mdtopdf` supports colourised output via the [gohighlight module](https://github.com/jessp01/gohighlight).

For examples, see `testdata/Markdown Documentation - Syntax.text` and `testdata/Markdown Documentation - Syntax.pdf`

## Custom themes

`md2pdf` supports both light and dark themes out of the box (use `--theme light` or `--theme dark` - no config required). 

However, if you wish to customise the font faces, sizes and colours, you can use the JSONs in
[custom_themes](./custom_themes) as a starting point. Edit to your liking and pass `--theme /path/to/json` to `md2pdf`

## Quick start

In the `cmd` folder is an example using the package. It demonstrates
a number of features. The test PDF was created with this command:
```
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

This repo has the [gohighlight module](https://github.com/jessp01/gohighlight) configured as a submodule so if you clone
with `--recursive`, you will have the `highlight` dir in its root. Alternatively, you may issue the below to update an
existing clone:

```sh
git submodule update --remote  --init
```

*Note 1: the `cmd` folder has an example for the syntax highlighting. 
See the script `run_syntax_highlighting.sh`. This example assumes that
the folder with the syntax files is located at relative location:
`../../../jessp01/gohighlight/syntax_files`.*

*Note 2: when annotating the code block to specify the language, the
annotation name must match syntax base filename.*

### Additional options

```sh
  -author string
    	Author; used if -footer is passed
  -font-file string
    	path to font file to use
  -font-name string
    	Font name ID; e.g 'Helvetica-1251'
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
    	Print doc footer (author  title  page number)
```

For example, the below will:

- Set the title to `My Grand Title`
- Set `Random Bloke` as the author (used in the footer)
- Set the dark theme
- Start a new page when encountering a HR (`---`); useful for creating presentations
- Print a footer (`author name, title, page number`)

```sh
$ go run md2pdf.go  -i /path/to/md \
    -o /path/to/pdf --title "My Grand Title" --author "Random Bloke" \
    --theme dark --new-page-on-hr --with-footer
```

## Using non-ASCII Glyphs/Fonts

In order to use a non-ASCII language there are a number things that must be done. The PDF generator must be configured with `WithUnicodeTranslator`:

```go
// https://en.wikipedia.org/wiki/Windows-1251
pf := mdtopdf.NewPdfRenderer("", "", *output, "trace.log", mdtopdf.WithUnicodeTranslator("cp1251")) 
```

In addition, this package's `Styler` must be used to set the font to match that is configured with the PDF generator.

A complete working example may be found for Russian in the `cmd` folder named
`russian.go`.

For a full example, run:

```sh
$ go run md2pdf.go -i russian.md -o russian.pdf \
    --unicode-encoding cp1251 --font-file helvetica_1251.json --font-name Helvetica_1251
```

## Tests

The tests included in this repo (see the `testdata` folder) were taken from the BlackFriday package.
They create PDF files and while the tests may complete
without errors, visual inspection of the created PDF is the
only way to determine if the tests *really* pass!

The tests create log files that trace the [gomarkdown](https://github.com/gomarkdown/markdown) parser
callbacks. This is a valuable debug tool showing each callback 
and data provided in each while the AST is presented.

## Limitations and Known Issues

- It is common for Markdown to include HTML. HTML is treated as a "code block". *There is no attempt to convert raw HTML to PDF.*
- Github-flavored Markdown permits strikethough using tildes. This is not supported at present by `fpdf` as a font style.
- The markdown link title, which would show when converted to HTML as hover-over text, is not supported. The generated PDF will show the actual URL that will be used if clicked, but this is a function of the PDF viewer.
- Definition lists are not supported (not sure that markdown supports them -- I need to research this)
- The following text features may be tweaked: font, size, spacing, style, fill color, and text color. These are exported and available via the `Styler` struct. Note that fill color only works when using `CellFormat()`. This is the case for: tables, codeblocks, and backticked text.


### Post release note 

In order to update `pkg.go.dev` with latest release, the following will do the trick. 
Essentially, it is creating a module and then running the go get command for the
desired release.
Using the proxy will have the side effect of updating the info on the go pkg web site.

```sh
$ pwd
/home/cecil/Downloads
$ mkdir tmp
$ cd tmp
$ ls
$ go mod init example.com/mypkg
go: creating new go.mod: module example.com/mypkg
$ cat go.mod 
module example.com/mypkg

go 1.20
$ GOPROXY=https://proxy.golang.org GO111MODULE=on go get github.com/solworktech/mdtopdf@v1.4.1
go: added github.com/go-pdf/fpdf v0.8.0
go: added github.com/jessp01/gohighlight v0.21.1-7
go: added github.com/solworktech/mdtopdf v1.4.1
go: added github.com/gomarkdown/markdown 
go: added gopkg.in/yaml.v2 v2.4.0
```

[license]: ./LICENSE
[badge-license]: https://img.shields.io/github/license/solworktech/mdtopdf.svg
[go-docs-badge]: https://godoc.org/github.com/mandolyte/mdtopdf?status.svg
[go-docs]: https://godoc.org/github.com/mandolyte/mdtopdf/v2
[badge-build]: https://github.com/solworktech/mdtopdf/actions/workflows/go.yml/badge.svg
[build]: https://github.com/solworktech/mdtopdf/actions/workflows/go.yml
[go-report-card-badge]: https://goreportcard.com/badge/github.com/mandolyte/mdtopdf/v2
[go-report-card]: https://goreportcard.com/report/github.com/mandolyte/mdtopdf/v2
