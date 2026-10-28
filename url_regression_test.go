package htmltomarkdown_test

import (
	"sync"
	"testing"

	"github.com/soulteary/html-to-markdown/v2/converter"
	"github.com/soulteary/html-to-markdown/v2/plugin/base"
	"github.com/soulteary/html-to-markdown/v2/plugin/commonmark"
)

func TestURLMarkerEncodingAcrossDocuments(t *testing.T) {
	cases := []struct{ name, input, domain, want string }{
		{"ordinary", `<a href="/a?x=1&amp;y=2">link</a><img src="/photo.png" alt="photo">`, "", `[link](/a?x=1&y=2)![photo](/photo.png)`},
		{"literal markers", `<a href="/a&#7;&#61442;?q=&#7;">link</a><img src="/b&#61442;&#7;.png">`, "", `[link](/a%07%EF%80%82?q=%07)![](/b%EF%80%82%07.png)`},
		{"encoded markers", `<a href="/a%07%EF%80%82">link</a><img src="/b%2507.png">`, "", `[link](/a%07%EF%80%82)![](/b%2507.png)`},
		{"relative URLs", `<a href="page&#7;">link</a><img src="image&#61442;.png">`, "https://example.com/docs/", `[link](https://example.com/docs/page%07)![](https://example.com/docs/image%EF%80%82.png)`},
		{"document-specific tokens", "<p>\ue000\a</p><a href='a\uf002'>link</a>", "https://other.example/", "\ue000\a\n\n[link](https://other.example/a%EF%80%82)"},
	}
	newConverter := func() *converter.Converter {
		return converter.NewConverter(converter.WithPlugins(base.NewBasePlugin(), commonmark.NewCommonmarkPlugin()))
	}
	shared := newConverter()
	check := func(c *converter.Converter) {
		// Interleave protected and ordinary documents to detect cross-document state.
		for repeat := 0; repeat < 3; repeat++ {
			for _, tc := range cases {
				got, err := c.ConvertString(tc.input, converter.WithDomain(tc.domain))
				if err != nil || got != tc.want {
					t.Errorf("%s: got %q, err=%v; want %q", tc.name, got, err, tc.want)
				}
			}
		}
	}
	check(shared)
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c := shared
			if i%2 == 0 {
				c = newConverter()
			}
			check(c)
		}(i)
	}
	wg.Wait()
}
