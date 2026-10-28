# Writing plugins

Create a converter with the base and CommonMark plugins, then register a handler
for an element name:

```go
conv := converter.NewConverter(converter.WithPlugins(
    base.NewBasePlugin(),
    commonmark.NewCommonmarkPlugin(),
))
conv.Register.RendererFor(
    "star-rating",
    converter.TagTypeInline,
    renderStarRating,
    converter.PriorityEarly,
)
```

A renderer receives the conversion context, an output writer, and the current
`*html.Node`. Return `converter.RenderSuccess` after writing output. Return
`converter.RenderTryNext` without writing when the handler cannot render that
node; later handlers can preserve its fallback content.

The complete [custom renderer example](examples/register/main.go) includes input
validation, attribute lookup, and fallback tests. Use public `html.Node` fields
and methods in external plugins; this repository's `internal/` helpers are not
part of the public API.

## Tag types and priorities

Choose `TagTypeInline` for content inside a paragraph, `TagTypeBlock` for block
content, and `TagTypeRemove` for elements to discard. `RendererFor` registers both
the tag type and the renderer. Use `PriorityEarly` when overriding an existing
renderer; `PriorityStandard` is appropriate for ordinary handlers. Priorities
with lower values run first.

The base plugin removes elements such as scripts and styles before rendering.
If a custom handler needs an element that is normally removed, register the
appropriate tag type early so normalization retains it.

Use `ctx.RenderChildNodes(ctx, w, node)` when a handler should apply normal
rendering to the current element's children. Rendered Markdown and raw text need
different escaping treatment; use the context's escaping methods when emitting
literal text. [ESCAPING.md](ESCAPING.md) explains the defaults.

## Package related handlers

A reusable plugin implements `converter.Plugin`:

```go
type ratingPlugin struct{}

func (ratingPlugin) Name() string { return "rating" }

func (ratingPlugin) Init(conv *converter.Converter) error {
    conv.Register.RendererFor(
        "star-rating", converter.TagTypeInline,
        renderStarRating, converter.PriorityEarly,
    )
    return nil
}
```

Register it with `converter.WithPlugins` alongside the base and CommonMark
plugins. Finish configuration before sharing a converter between goroutines, and
keep custom state safe for concurrent conversions. Add tests for valid input,
malformed attributes, fallback content, and interactions with built-in renderers.
