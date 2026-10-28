package base

import (
	"bytes"
	"strings"

	"github.com/soulteary/html-to-markdown/v2/collapse"
	"github.com/soulteary/html-to-markdown/v2/converter"
	"github.com/soulteary/html-to-markdown/v2/internal/domutils"
	"github.com/soulteary/html-to-markdown/v2/internal/textutils"
	"golang.org/x/net/html"
)

type base struct{}

// NewBasePlugin registers a bunch of stuff that is not necessarily related to commonmark,
// like removing nodes, trimming whitespace, collapsing whitespace, ...
func NewBasePlugin() converter.Plugin {
	base := base{}
	return &base
}

func (s *base) Name() string {
	return "base"
}
func (b *base) Init(conv *converter.Converter) error {
	conv.Register.TagType("#comment", converter.TagTypeRemove, converter.PriorityStandard)
	conv.Register.TagType("head", converter.TagTypeRemove, converter.PriorityStandard)
	conv.Register.TagType("script", converter.TagTypeRemove, converter.PriorityStandard)
	conv.Register.TagType("style", converter.TagTypeRemove, converter.PriorityStandard)
	conv.Register.TagType("link", converter.TagTypeRemove, converter.PriorityStandard)
	conv.Register.TagType("meta", converter.TagTypeRemove, converter.PriorityStandard)

	conv.Register.TagType("iframe", converter.TagTypeRemove, converter.PriorityStandard)
	conv.Register.TagType("noscript", converter.TagTypeRemove, converter.PriorityStandard)

	conv.Register.TagType("input", converter.TagTypeRemove, converter.PriorityStandard)
	conv.Register.TagType("textarea", converter.TagTypeRemove, converter.PriorityStandard)

	// "tr" is not in the `IsBlockNode` list,
	// but we want to treat is as a block anyway.
	// conv.Register.TagStrategy("tr", converter.StrategyMarkdownBlock, converter.PriorityStandard)
	// conv.Register.TagType("tr", converter.BlockTagType, converter.PriorityStandard)

	conv.Register.PreRenderer(b.preRenderRemove, converter.PriorityEarly)
	// Note: The priority is low, so that collapse runs _after_ all the other functions
	conv.Register.PreRenderer(b.preRenderCollapse, converter.PriorityLate)

	conv.Register.TextTransformer(b.handleTextTransform, converter.PriorityStandard)

	conv.Register.PostRenderer(b.postRenderTrimContent, converter.PriorityStandard)
	conv.Register.PostRenderer(b.postRenderUnescapeContent, converter.PriorityStandard+20)

	return nil
}

func (b *base) preRenderRemove(ctx converter.Context, doc *html.Node) {
	var finder func(node *html.Node)
	finder = func(node *html.Node) {
		name := domutils.NodeName(node)

		if tagType, _ := ctx.GetTagType(name); tagType == converter.TagTypeRemove {
			domutils.RemoveNode(node)
			return
		}

		// Match the original deferred walk's right-to-left order without
		// allocating a deferred call for every child. Save the neighbor before
		// recursion, since the current child may be detached.
		for child := node.LastChild; child != nil; {
			previous := child.PrevSibling
			finder(child)
			child = previous
		}
	}
	finder(doc)

	// - - - - - - - //

	// After removing elements (see above) it can happen that we have
	// two #text nodes right next to each other. This would cause problems
	// with the collapse so we merge them together.
	domutils.MergeAdjacentTextNodes(doc)
}

func (b *base) preRenderCollapse(ctx converter.Context, doc *html.Node) {
	collapse.Collapse(doc, &collapse.DomFuncs{
		IsBlockNode: func(node *html.Node) bool {
			tagName := domutils.NodeName(node)
			tagType, ok := ctx.GetTagType(tagName)
			if ok {
				return tagType == converter.TagTypeBlock
			}

			return domutils.NameIsBlockNode(tagName)
		},
	})
}

var characterEntityReplacer = strings.NewReplacer(
	// We are not using `html.EscapeString` because we
	// care about fewer characters
	"<", "&lt;",
	">", "&gt;",

	// Note: We are not escaping "&" as "&amp;" anymore.
	// In most cases the "&" is completely fine.
	// https://github.com/JohannesKaufmann/html-to-markdown/issues/178
)

func (b *base) handleTextTransform(ctx converter.Context, content string) string {

	// TODO: similar to UnEscapers also only escape if nessesary.
	//       "<" only if not followed by space
	//       "&" only if character entity
	content = characterEntityReplacer.Replace(content)

	// TODO: reduce conversion between types
	content = string(ctx.EscapeContent([]byte(content)))

	return content
}

func (b *base) postRenderTrimContent(ctx converter.Context, result []byte) []byte {
	// Remove whitespace from the beginning & end
	result = bytes.TrimSpace(result)

	// Remove too many newlines
	result = textutils.TrimConsecutiveNewlines(result)
	result = textutils.TrimUnnecessaryHardLineBreaks(result)

	return result
}
func (b *base) postRenderUnescapeContent(ctx converter.Context, result []byte) []byte {
	result = ctx.UnEscapeContent(result)
	return result
}
