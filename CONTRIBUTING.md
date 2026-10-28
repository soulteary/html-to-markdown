# Contributing

Use Go 1.25 or newer. The module path is
`github.com/soulteary/html-to-markdown/v2`.

## Repository layout

- Root package: string, reader, and parsed-node conversion APIs.
- `converter/`: configuration, handler registration, context, and rendering.
- `plugin/`: base, CommonMark, tables, and strikethrough.
- `collapse/`: whitespace normalization.
- `internal/`: implementation and test helpers.
- `examples/`: runnable programs with executable output tests.

## Validate changes

```sh
gofmt -w .
go build ./...
go vet ./...
go test -race ./... -count=1
```

Keep rendering changes covered by a focused regression test or a golden fixture.
Run `go test ./examples/...` when editing examples; also run a changed example
with `go run ./examples/<name>` before updating its documented output.

## Golden fixtures

Plugin golden tests pair `.in.html` and `.out.md` files under each plugin's
`testdata/` directory. Normal tests never overwrite fixtures. To intentionally
update them:

```sh
go test ./plugin/commonmark ./plugin/strikethrough ./plugin/table -update -count=1
```

Review every fixture change in Git. A changed expectation is not a substitute
for checking that the new rendering is correct. Add small regression cases for
edge conditions, including escaping, whitespace, and nested blocks.

## Dependencies and attribution

Keep runtime dependencies focused on the library's needs. Run `go mod tidy`
after dependency changes. Preserve the original MIT license and the attribution
in `internal/domutils/LICENSE.dom` when modifying adapted DOM helpers.

For bug reports, include the module version, Go version, enabled plugins,
minimal Go code, HTML input, actual Markdown, and expected Markdown. See
[SECURITY.md](SECURITY.md) for vulnerability reporting.

## Library releases

Use a new `v2.x.y` tag for a library release after the validation commands pass.
The tagged commit must declare the `github.com/soulteary/html-to-markdown/v2`
module path. Historical tags inherited from upstream can still declare its old
module path, so the README installs `@main` during this transition.
