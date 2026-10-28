package converter

import (
	"errors"
	"fmt"

	"github.com/soulteary/html-to-markdown/v2/internal/domutils"
	"golang.org/x/net/html"
)

type register struct {
	conv *Converter
}

func (r *register) Plugin(plugin Plugin) {
	pluginName := plugin.Name()
	if pluginName == "" {
		r.conv.setError(errors.New("the plugin has no name"))
		return
	}

	r.conv.m.Lock()
	r.conv.registeredPlugins = append(r.conv.registeredPlugins, pluginName)
	r.conv.m.Unlock()

	err := plugin.Init(r.conv)
	if err != nil {
		r.conv.setError(fmt.Errorf("error while initializing %q plugin: %w", pluginName, err))
		return
	}
}

// - - - - - - - - - - - - - Pre-Render - - - - - - - - - - - - - //

type HandlePreRenderFunc func(ctx Context, doc *html.Node)

func (r *register) PreRenderer(fn HandlePreRenderFunc, priority int) {
	r.conv.m.Lock()
	defer r.conv.m.Unlock()

	handler := prioritized(fn, priority)
	r.conv.preRenderHandlers.add(handler)
}
func (conv *Converter) getPreRenderHandlers() prioritizedSlice[HandlePreRenderFunc] {
	return sortedHandlers(&conv.m, &conv.preRenderHandlers)
}

// - - - - - - - - - - - - - Render - - - - - - - - - - - - - //

// Writer is an interface that only conforms to the Write* methods of bytes.Buffer
type Writer interface {
	Write(p []byte) (n int, err error)
	WriteByte(c byte) error
	WriteRune(r rune) (n int, err error)
	WriteString(s string) (n int, err error)
}

type HandleRenderFunc func(ctx Context, w Writer, n *html.Node) RenderStatus

func (r *register) Renderer(fn HandleRenderFunc, priority int) {
	r.conv.m.Lock()
	defer r.conv.m.Unlock()

	handler := prioritized(fn, priority)
	r.conv.renderHandlers.add(handler)
}

// RendererFor registers a renderer for a specific tag (e.g. "div").
// It is a small wrapper around `TagType()` and `Renderer()`.
func (r *register) RendererFor(tagName string, tagType tagType, renderFn HandleRenderFunc, priority int) {

	// 1. we add the "tagType" to the map
	r.TagType(tagName, tagType, priority)

	// 2. we register the render function
	r.Renderer(func(ctx Context, w Writer, n *html.Node) RenderStatus {
		if domutils.NodeName(n) == tagName {
			return renderFn(ctx, w, n)
		}
		return RenderTryNext
	}, priority)
}

func (conv *Converter) getRenderHandlers() prioritizedSlice[HandleRenderFunc] {
	return sortedHandlers(&conv.m, &conv.renderHandlers)
}

// - - - - - - - - - - - - - Post Render - - - - - - - - - - - - - //

type HandlePostRenderFunc func(ctx Context, content []byte) []byte

func (r *register) PostRenderer(fn HandlePostRenderFunc, priority int) {
	r.conv.m.Lock()
	defer r.conv.m.Unlock()

	handler := prioritized(fn, priority)
	r.conv.postRenderHandlers.add(handler)
}
func (conv *Converter) getPostRenderHandlers() prioritizedSlice[HandlePostRenderFunc] {
	return sortedHandlers(&conv.m, &conv.postRenderHandlers)
}

// - - - - - - - - - - - - - Text - - - - - - - - - - - - - //

type HandleTextTransformFunc func(ctx Context, content string) string

func (r *register) TextTransformer(fn HandleTextTransformFunc, priority int) {
	r.conv.m.Lock()
	defer r.conv.m.Unlock()

	handler := prioritized(fn, priority)
	r.conv.textTransformHandlers.add(handler)
}
func (conv *Converter) getTextTransformHandlers() prioritizedSlice[HandleTextTransformFunc] {
	return sortedHandlers(&conv.m, &conv.textTransformHandlers)
}

// - - - - - - - - - - - - - Escaping - - - - - - - - - - - - - //

func (r *register) EscapedChar(chars ...rune) {
	r.conv.m.Lock()
	defer r.conv.m.Unlock()

	for _, char := range chars {
		r.conv.markdownChars[char] = struct{}{}
	}
}

type HandleUnEscapeFunc func(chars []byte, index int) int

func (r *register) UnEscaper(fn HandleUnEscapeFunc, priority int) {
	r.conv.m.Lock()
	defer r.conv.m.Unlock()

	handler := prioritized(fn, priority)
	r.conv.unEscapeHandlers.add(handler)
}
func (conv *Converter) getUnEscapeHandlers() prioritizedSlice[HandleUnEscapeFunc] {
	return sortedHandlers(&conv.m, &conv.unEscapeHandlers)
}

// - - - - - - - - - - - - - Tag Type - - - - - - - - - - - - - //

type tagType string

const (
	TagTypeBlock  tagType = "block"
	TagTypeInline tagType = "inline"

	// TagTypeRemove will remove that node in the _PreRender_ phase with a high priority.
	TagTypeRemove tagType = "remove"
)

func (r *register) TagType(tagName string, kind tagType, priority int) {
	r.conv.m.Lock()
	defer r.conv.m.Unlock()

	val := prioritized(kind, priority)
	// Publish a new sorted slice so readers never sort shared storage.
	types := append(prioritizedSlice[tagType](nil), r.conv.tagTypes[tagName]...)
	types = append(types, val)
	types.Sort()
	r.conv.tagTypes[tagName] = types
}
func (conv *Converter) getTagType(tagName string) (tagType, bool) {
	conv.m.RLock()
	defer conv.m.RUnlock()

	types, ok := conv.tagTypes[tagName]
	if !ok || len(types) == 0 {

		if domutils.NameIsBlockNode(tagName) {
			return TagTypeBlock, true
		} else if domutils.NameIsInlineNode(tagName) {
			return TagTypeInline, true
		}
		return "", false
	}

	firstType := types[0].Value

	return firstType, true
}
