package comparison_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func TestInputOutputParity(t *testing.T) {
	cases := []struct {
		name, input string
		tables      bool
	}{
		{"Small", `<h1>Hello</h1><p>A <strong>small</strong> example.</p>`, false},
		{"Article", strings.Repeat(`<section><h2>Go library</h2><p>HTML conversion with <strong>formatting</strong> and useful plain text.</p></section>`, 200), false},
		{"MarkdownText", `<p>` + strings.Repeat("中文 *word* _x_ [label](path) ! # + - 1. ~ `code` &lt;tag&gt; ", 500) + `</p>`, false},
		{"Links", strings.Repeat(`<p><a href="https://example.com/a?x=1&amp;y=2">A link</a> <img src="/photo.png" alt="photo"></p>`, 200), false},
		{"ManyNodes", `<div>` + strings.Repeat(`<span>text</span>`, 2000) + `</div>`, false},
		{"WideRemovals", `<div>` + strings.Repeat(`<span>kept</span><script>ignored</script><style>ignored</style>`, 1000) + `</div>`, false},
		{"LongUnicodeText", `<p>` + strings.Repeat("中文 text with   spaces and punctuation. ", 2000) + `</p>`, false},
		{"DeepTree", strings.Repeat(`<div>`, 128) + `<p>deep text</p>` + strings.Repeat(`</div>`, 128), false},
		{"Table", `<table><tr><th>Name</th><th>Value</th></tr>` + strings.Repeat(`<tr><td><strong>Item</strong></td><td>42</td></tr>`, 200) + `</table>`, true},
	}
	for _, tc := range cases {
		u, ue := upstreamConverter(tc.tables).ConvertString(tc.input)
		f, fe := forkConverter(tc.tables).ConvertString(tc.input)
		if ue != nil || fe != nil || u != f {
			t.Fatalf("output mismatch for %s: %v, %v", tc.name, ue, fe)
		}
		row := map[string]any{"case": tc.name, "input_bytes": len(tc.input), "input_sha256": fmt.Sprintf("%x", sha256.Sum256([]byte(tc.input))), "output_bytes": len(u), "upstream_output_sha256": fmt.Sprintf("%x", sha256.Sum256([]byte(u))), "fork_output_sha256": fmt.Sprintf("%x", sha256.Sum256([]byte(f)))}
		b, _ := json.Marshal(row)
		fmt.Println("PARITY " + string(b))
	}
	fmt.Printf("ENV go=%s os=%s arch=%s gomaxprocs=%d cpu_count=%d\n", runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.GOMAXPROCS(0), runtime.NumCPU())
}
