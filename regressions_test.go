package htmltomarkdown_test

import (
	"os"
	"strings"
	"sync"
	"testing"

	htmltomarkdown "github.com/soulteary/html-to-markdown/v2"
	"github.com/soulteary/html-to-markdown/v2/converter"
	"github.com/soulteary/html-to-markdown/v2/plugin/base"
	"github.com/soulteary/html-to-markdown/v2/plugin/commonmark"
	"github.com/soulteary/html-to-markdown/v2/plugin/table"
	"golang.org/x/net/html"
)

func TestUpstreamConversionRegressions(t *testing.T) {
	cases := []struct{ name, input, want string }{
		{"223 bell in text", "<p>a&#7;b</p>", "a\ab"},
		{"223 trailing bell", "<p>a&#7;</p>", "a\a"},
		{"223 bell before emphasis", "<p>&#7;<strong>bold</strong></p>", "\a**bold**"},
		{"223 consecutive bells", "<p>&#7;&#7;*</p>", "\a\a*"},
		{"223 private use marker", "<p>a\uf002b</p>", "a\uf002b"},
		{"223 code markers", "<pre><code>a\a\uf002b\nc</code></pre>", "```\na\a\uf002b\nc\n```"},
		{"223 inline markers", "<code>a\a\uf002b</code>", "`a\a\uf002b`"},
		{"223 existing token candidates", "<p>\ue000\ue001\ue000\ue000\ue002\a\uf002</p>", "\ue000\ue001\ue000\ue000\ue002\a\uf002"},
		{"223 image attributes", "<img src='a\uf002.png' alt='x\ay'>", "![x\ay](a%EF%80%82.png)"},
		{"223 bell in image URL", "<img src='a&#7;.png'>", "![](a%07.png)"},
		{"223 encoded token candidates", "<p>%ee%80%80%ee%80%82\a\uf002</p>", "%ee%80%80%ee%80%82\a\uf002"},
		{"223 adjacent input token fragments", "<span>\ue000</span><span>\ue000</span><span>\ue001</span><span>\a</span>", "\ue000\ue000\ue001\a"},
		{"223 encoded URL query", "<a href='/a?q=\uf002'>link</a>", "[link](/a?q=%EF%80%82)"},
		{"183 quoted code", "<blockquote><pre>line 1\nline 2\n</pre></blockquote>", "> ```\n> line 1\n> line 2\n> ```"},
		{"183 empty code lines", "<blockquote><pre>a\n\nb\n</pre></blockquote>", "> ```\n> a\n> \n> b\n> ```"},
		{"183 nested quote", "<blockquote><blockquote><pre>a\nb</pre></blockquote></blockquote>", "> > ```\n> > a\n> > b\n> > ```"},
		{"183 quote in list", "<ul><li><blockquote><pre>a\nb</pre></blockquote></li></ul>", "- > ```\n  > a\n  > b\n  > ```"},
		{"224 adjacent code", "<pre><code>c1</code><code>c2</code></pre>", "```\nc1\nc2\n```"},
		{"224 empty code line", "<pre><code>c1</code><code></code><code>c2</code></pre>", "```\nc1\n\nc2\n```"},
		{"224 existing newlines", "<pre><code>c1\n</code><code>c2\n</code></pre>", "```\nc1\nc2\n```"},
		{"224 highlighted code", "<pre data-lang='javascript'><code><span>let a = 1;</span></code><code><span>let b = 2;</span></code></pre>", "```javascript\nlet a = 1;\nlet b = 2;\n```"},
		{"224 language attribute", "<pre data-language='go'><code>one</code><code>two</code></pre>", "```go\none\ntwo\n```"},
		{"224 inline unchanged", "<p><code>a</code><code>b</code></p>", "`ab`"},
		{"199 original attribute", `<img data-original="/original.png" alt="[alt]" title="title">`, "![\\[alt\\]](/original.png \"title\")"},
		{"199 data src", `<img data-src="/lazy.png">`, "![](/lazy.png)"},
		{"199 lazy src", `<img data-lazy-src="/lazy.png">`, "![](/lazy.png)"},
		{"199 src priority", `<img src="/normal.png" data-src="/lazy.png" data-original="/original.png">`, "![](/normal.png)"},
		{"199 blank fallback", `<img src="  " data-src=" " data-original=" /original.png ">`, "![](/original.png)"},
		{"199 no guessed attribute", `<img data-tracking-url="/tracking.png">`, ""},
		{"185 gutter table", `<table><tbody><tr><td class="gutter"><pre><span class="line">1</span><br><span class="line">2</span><br></pre></td><td class="code"><pre><span class="line">a</span><br><span class="line">  b</span><br></pre></td></tr></tbody></table>`, "```\na\n  b\n```"},
		{"185 language and class tokens", `<table><tr><td class="other gutter"><pre>5<br>6<br></pre></td><td class="code other"><pre class="language-go">a<br>b<br></pre></td></tr></table>`, "```go\na\nb\n```"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := htmltomarkdown.ConvertString(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("got  %q\nwant %q", got, tc.want)
			}
		})
	}
}

func TestLiteralMarkersRepeatedConvertNode(t *testing.T) {
	doc, err := html.Parse(strings.NewReader("<p>a\a\uf002b</p>"))
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		out, err := htmltomarkdown.ConvertNode(doc)
		if err != nil {
			t.Fatal(err)
		}
		if string(out) != "a\a\uf002b" {
			t.Fatalf("conversion %d: %q", i, out)
		}
	}
}

func TestLazyImageDomain(t *testing.T) {
	out, err := htmltomarkdown.ConvertString(`<img data-original="/image.png">`, converter.WithDomain("https://example.com"))
	if err != nil {
		t.Fatal(err)
	}
	if out != "![](https://example.com/image.png)" {
		t.Fatal(out)
	}
}

func TestCodeTableRecognition(t *testing.T) {
	codeTable := `<table><tr><td class="gutter"><pre>1<br>2<br></pre></td><td class="code"><pre>a<br>b<br></pre></td></tr></table>`
	for _, preserve := range []bool{false, true} {
		conv := converter.NewConverter(converter.WithPlugins(base.NewBasePlugin(), commonmark.NewCommonmarkPlugin(), table.NewTablePlugin(table.WithNewlineBehavior(map[bool]table.NewlineBehavior{false: table.NewlineBehaviorSkip, true: table.NewlineBehaviorPreserve}[preserve]))))
		out, err := conv.ConvertString(codeTable)
		if err != nil {
			t.Fatal(err)
		}
		if out != "```\na\nb\n```" {
			t.Fatalf("preserve=%v: %q", preserve, out)
		}
	}
	// A matching class alone must not discard meaningful tabular content.
	for _, input := range []string{
		`<table><tr><td class="gutter"><pre>name<br>value</pre></td><td class="code"><pre>a<br>b</pre></td></tr></table>`,
		`<table><caption>Important caption</caption><tr><td class="gutter"><pre>1<br>2</pre></td><td class="code"><pre>a<br>b</pre></td></tr></table>`,
		`<table><tr><td class="gutter"><pre>1<br>3</pre></td><td class="code"><pre>a<br>b</pre></td></tr></table>`,
		`<table><tr><td class="gutter"><pre>1</pre></td><td class="code"><pre>a<br>b</pre></td></tr></table>`,
		`<table><tr><td class="gutter"><pre>1<br>2</pre></td><td class="code">Keep me<pre>a<br>b</pre></td></tr></table>`,
	} {
		out, err := htmltomarkdown.ConvertString(input)
		if err != nil {
			t.Fatal(err)
		}
		if out == "```\na\nb\n```" {
			t.Fatalf("incorrectly discarded table content: %s", input)
		}
	}
}

func TestOriginalIssue185(t *testing.T) {
	input, err := os.ReadFile("testdata/regressions/issue185.html")
	if err != nil {
		t.Fatal(err)
	}
	want := "```\nBOOL AddPrinterDriverEx(\n  _In_    LPTSTR pName,\n  _In_    DWORD  Level,\n  _Inout_ LPBYTE pDriverInfo,\n  _In_    DWORD  dwFileCopyFlags\n);\n```"
	got, err := htmltomarkdown.ConvertString(string(input))
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %q; want %q", got, want)
	}
}

func TestCodeInOrdinaryTableFallsBack(t *testing.T) {
	conv := converter.NewConverter(converter.WithPlugins(base.NewBasePlugin(), commonmark.NewCommonmarkPlugin(), table.NewTablePlugin(table.WithNewlineBehavior(table.NewlineBehaviorPreserve))))
	got, err := conv.ConvertString(`<table><tr><td><pre>a
b</pre></td></tr></table>`)
	if err != nil {
		t.Fatal(err)
	}
	if got != "```\na\nb\n```" {
		t.Fatalf("fenced block must not be embedded in a pipe table: %q", got)
	}
}

func TestLiteralMarkersEscapeModes(t *testing.T) {
	for _, disabled := range []bool{false, true} {
		mode := converter.EscapeModeSmart
		if disabled {
			mode = converter.EscapeModeDisabled
		}
		conv := converter.NewConverter(converter.WithEscapeMode(mode), converter.WithPlugins(base.NewBasePlugin(), commonmark.NewCommonmarkPlugin()))
		got, err := conv.ConvertString("<ul><li>a\a\uf002b</li></ul>")
		if err != nil {
			t.Fatal(err)
		}
		if got != "- a\a\uf002b" {
			t.Fatalf("disabled=%v: %q", disabled, got)
		}
	}
}

func TestConcurrentLiteralMarkers(t *testing.T) {
	conv := converter.NewConverter(converter.WithPlugins(base.NewBasePlugin(), commonmark.NewCommonmarkPlugin()))
	var workers sync.WaitGroup
	for i := 0; i < 16; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for j := 0; j < 20; j++ {
				got, err := conv.ConvertString("<pre>\ue000\a\uf002\nx</pre>")
				if err != nil || got != "```\n\ue000\a\uf002\nx\n```" {
					t.Errorf("got %q, err=%v", got, err)
					return
				}
			}
		}()
	}
	workers.Wait()
}

func TestMarkerTokensExcludeBaseURL(t *testing.T) {
	for _, domain := range []string{"https://example.com/%EE%80%80/", "https://example.com/\ue000/", "https://example.com/%EE%80%81/"} {
		got, err := htmltomarkdown.ConvertString("<p>&#7;&#61442;</p><a href='page'>link</a><img data-original='image.png'>", converter.WithDomain(domain))
		if err != nil {
			t.Fatal(err)
		}
		encoded := strings.ReplaceAll(domain, "\ue000", "%EE%80%80")
		want := "\a\uf002\n\n[link](" + encoded + "page)![](" + encoded + "image.png)"
		if got != want {
			t.Errorf("base %q: got %q; want %q", domain, got, want)
		}
	}
}

func TestMarkerTokensAvoidInternalMarkers(t *testing.T) {
	var text strings.Builder
	for r := rune(0xe000); r < 0xf002; r++ {
		text.WriteRune(r)
	}
	text.WriteRune('\a')
	want := text.String()
	got, err := htmltomarkdown.ConvertString("<p>" + want + "</p>")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("literal marker corrupted after %d private-use runes", 0xf002-0xe000)
	}
}

func TestEncodedMarkerTextAcrossInlineNodes(t *testing.T) {
	for _, input := range []string{
		`<p>&#7;</p><span>%EE</span><span>%80%80</span>`,
		`<p>&#7;</p><span>%EE</span><script>ignored</script><span>%80%80</span>`,
	} {
		got, err := htmltomarkdown.ConvertString(input)
		if err != nil {
			t.Fatal(err)
		}
		if got != "\a\n\n%EE%80%80" {
			t.Fatalf("got %q", got)
		}
	}
}
