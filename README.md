# html-to-markdown

A Go library for converting HTML to Markdown, with CommonMark rendering,
configurable plugins, and custom element renderers.

This is an independently maintained fork of
[JohannesKaufmann/html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown).
The module path is `github.com/soulteary/html-to-markdown/v2`.
Go 1.25 or newer is required. The only direct external dependency is
`golang.org/x/net`, used for HTML parsing.

## Install

Install the current maintained branch:

```sh
go get github.com/soulteary/html-to-markdown/v2@main
```

When migrating from upstream, change the `github.com/JohannesKaufmann/html-to-markdown/v2`
import prefix to `github.com/soulteary/html-to-markdown/v2`, then run `go mod tidy`.

## Quick start

```go
package main

import (
    "fmt"
    "log"

    htmltomarkdown "github.com/soulteary/html-to-markdown/v2"
)

func main() {
    markdown, err := htmltomarkdown.ConvertString(
        `<h1>Hello</h1><p>A <strong>small</strong> example.</p>`,
    )
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(markdown)
}
```

Output:

```markdown
# Hello

A **small** example.
```

## Conversion APIs

- `ConvertString` accepts an HTML string and returns a Markdown string.
- `ConvertReader` accepts an `io.Reader` and returns Markdown bytes.
- `ConvertNode` accepts a parsed `*html.Node` and returns Markdown bytes.

These functions enable the base and CommonMark plugins. Use
`converter.WithDomain("https://example.com/docs/")` to resolve relative links and
image URLs against a document URL. The [reader example](examples/reader) shows this.

`ConvertNode` normalizes and mutates its input tree. Use separate trees for
concurrent conversions. A configured `converter.Converter` can be shared for
conversions; finish registering handlers before using it concurrently, and ensure
custom handlers are safe for that use.

## Configure rendering and plugins

Create a converter to configure Markdown formatting or enable optional plugins:

```go
conv := converter.NewConverter(
    converter.WithPlugins(
        base.NewBasePlugin(),
        commonmark.NewCommonmarkPlugin(
            commonmark.WithStrongDelimiter("__"),
        ),
        strikethrough.NewStrikethroughPlugin(),
        table.NewTablePlugin(
            table.WithCellPaddingBehavior(table.CellPaddingBehaviorNone),
        ),
    ),
)
markdown, err := conv.ConvertString(htmlInput)
```

Import `converter` and the plugins from this module's `/converter`,
`/plugin/base`, `/plugin/commonmark`, `/plugin/strikethrough`, and `/plugin/table`
packages. See the complete [options](examples/options) and [plugins](examples/plugins)
programs for runnable code.

The table plugin supports alignment, row/column spans, header promotion, empty-row
filtering, and compact cell padding. For simple line breaks in cells, configure
`table.WithNewlineBehavior(table.NewlineBehaviorPreserve)` to emit `<br />`.
Preformatted blocks fall back to block output when they cannot form a valid
Markdown table.

Images use `src`, falling back to `data-src`, `data-original`, and `data-lazy-src`
when earlier attributes are empty. Recognized syntax-highlighter gutter/code
layouts become one code block without line numbers.

For custom tags or rendering overrides, use `conv.Register.RendererFor`.
[Writing plugins](WRITING_PLUGINS.md) explains priorities, fallback behavior, and
how to package handlers. [Escaping](ESCAPING.md) explains smart escaping and its
configuration.

## Runnable examples

Each example is a small program with tested output:

- [Basics](examples/basics): HTML strings, headings, emphasis, and links.
- [Reader](examples/reader): `io.Reader` input and relative URL resolution.
- [Options](examples/options): emphasis delimiters and list markers.
- [Plugins](examples/plugins): tables and strikethrough together.
- [Custom renderer](examples/register): a custom tag with validated attributes and fallback text.

```sh
go run ./examples/basics
go run ./examples/reader
go run ./examples/options
go run ./examples/plugins
go run ./examples/register
```

See the [examples guide](examples/README.md) for expected outputs, and the
package's tested `Example` functions for the string, reader, and parsed-node APIs.

## Development

```sh
go build ./...
go vet ./...
go test -race ./...
```

Golden fixtures and regression cases cover the converter and plugins. Fixtures
are compared byte-for-byte; updates are explicit and should be reviewed.
[Contributing](CONTRIBUTING.md) documents tests, fixture updates, and repository
layout. CI builds and tests the library and examples on Linux, macOS, and Windows
with the minimum supported Go version and the current stable version.

## Security and license

Conversion does not sanitize untrusted content. Sanitize it before rendering
Markdown as HTML in a browser. See [SECURITY.md](SECURITY.md) for reporting.

[MIT licensed](LICENSE). Original copyright notices and the
[adapted DOM helper license](internal/domutils/LICENSE.dom) are retained.
