// Package htmltomarkdown converts HTML strings, readers, and parsed nodes to
// Markdown using the base and CommonMark plugins.
//
// Use the converter package to configure rendering or add optional plugins,
// including tables and strikethrough. See the examples directory for runnable
// programs and the Example functions for tested API usage.
//
// Conversion does not sanitize untrusted content. ConvertNode mutates the
// supplied HTML tree during normalization.
package htmltomarkdown
