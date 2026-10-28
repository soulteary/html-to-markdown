# Examples

Run these programs from the repository root. Each program has an executable
`Example` test; `go test ./examples/...` checks the documented output.
The programs use only public packages and can be copied into another module.

## Basic conversion

```sh
go run ./examples/basics
```

```markdown
# Library first

A **small** example with [a link](https://example.com).
```

[Source](basics/main.go) · [Output test](basics/main_test.go)

## Reader input and base URL

```sh
go run ./examples/reader
```

```markdown
[Guide](https://example.com/docs/guide) ![Logo](https://example.com/logo.png)
```

`ConvertReader` accepts any `io.Reader`, including an open file or HTTP response
body. The caller owns and closes those resources. `WithDomain` resolves relative
paths against the source document's URL.

[Source](reader/main.go) · [Output test](reader/main_test.go)

## Formatting options

```sh
go run ./examples/options
```

```markdown
__Custom formatting__

* First
* Second
```

A configured converter enables the base and CommonMark plugins and changes the
strong-emphasis delimiter and bullet marker.

[Source](options/main.go) · [Output test](options/main_test.go)

## Tables and strikethrough

```sh
go run ./examples/plugins
```

```markdown
|Name|Status|
|---|---|
|Legacy|~~active~~|
```

Optional plugins compose with the default rendering plugins. This example also
removes visual cell padding for compact Markdown.

[Source](plugins/main.go) · [Output test](plugins/main_test.go)

## Custom element renderer

```sh
go run ./examples/register
```

```markdown
## Go ★★★★
```

The `star-rating` handler reads public `html.Node` attributes, accepts counts
from zero to five, and returns `RenderTryNext` for invalid attributes so the
original fallback text survives. It does not import an internal helper package.

[Source](register/main.go) · [Output and fallback tests](register/main_test.go)

For parsed-node input, see `ExampleConvertNode` in [convert_test.go](../convert_test.go).
