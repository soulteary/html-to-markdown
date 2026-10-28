# Escaping

Markdown characters can change the meaning of literal text. For example,
`<p>**literal**</p>` contains plain asterisks, while
`<p><strong>literal</strong></p>` contains emphasis. Smart escaping preserves
that distinction by inserting backslashes where Markdown would otherwise
interpret literal characters as syntax.

Smart escaping is enabled by default. To explicitly select it on a configured
converter:

```go
conv := converter.NewConverter(
    converter.WithEscapeMode(converter.EscapeModeSmart),
    converter.WithPlugins(
        base.NewBasePlugin(),
        commonmark.NewCommonmarkPlugin(),
    ),
)
```

`converter.EscapeModeDisabled` passes Markdown-like text through without that
protection. Use it only when that interpretation is intended; literal `**text**`
can then render as emphasis. This setting does not sanitize untrusted content.

For custom renderers, distinguish Markdown syntax you intentionally generate
from literal text. Use the conversion context's escaping methods for literal
text, and add regression tests around punctuation, whitespace, and nested
formatting. Do not emit the library's internal marker characters as a substitute
for these methods.

If escaping produces incorrect output, open a bug report in this repository with
a minimal HTML input, converter configuration, actual Markdown, and expected
Markdown.
