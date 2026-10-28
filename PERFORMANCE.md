# Performance

Use Go 1.27.3 or newer. The end-to-end benchmarks in
`performance_benchmark_test.go` cover small fragments, repeated article
sections, dense Markdown punctuation with Unicode, links/images, many sibling
nodes, and tables with formatted cells. They include HTML parsing, DOM
normalization, rendering, escaping, and result allocation.

Each workload has two modes:

- `Fresh` constructs a converter for every conversion. CommonMark workloads use
  the root convenience API; the table workload also registers the table plugin.
- `Reuse` configures one converter outside the timed loop, then converts a fresh
  HTML string on every iteration. It still parses a new DOM each time.

Run repeated samples without the race detector when measuring performance:

```sh
go test -run '^$' -bench BenchmarkConvert -benchmem -benchtime=1s -count=10
```

Compare changes with the same Go version, hardware, inputs, and process settings.
Use `GOTOOLCHAIN=go1.27.3` for both revisions when comparing this upgrade with
an older revision. Copy the benchmark file to an older revision if it predates
these benchmarks. Interleave before/after runs and inspect allocation counts as
well as timings; scheduling, garbage collection, and thermal changes can affect
short runs. Synthetic workload results are not a guarantee for arbitrary pages.

For a CPU and allocation profile of a specific workload:

```sh
go test -run '^$' -bench 'BenchmarkConvert/MarkdownText/Reuse$' -benchtime=5s -cpuprofile=cpu.pprof -memprofile=memory.pprof
go tool pprof cpu.pprof
go tool pprof -alloc_space memory.pprof
```

Keep profiles outside version control. Run `go test -race ./...` separately to
validate correctness and concurrent use.

## Batch conversion

For repeated conversions, create a `converter.Converter` with the base and
CommonMark plugins once, then call its `ConvertString` or `ConvertReader`
methods for each document. Optional plugins and custom handlers can be added
during configuration. Finish registration before sharing the converter between
goroutines, and ensure custom handlers are safe for concurrent calls.

The root `ConvertString`, `ConvertReader`, and `ConvertNode` helpers construct a
new converter on each call. They are convenient for occasional conversions;
reusing a configured converter avoids repeated plugin registration and handler
snapshot construction in a batch. `ConvertNode` mutates its input, so each
concurrent conversion must receive a separate tree.

## Handler snapshots and escaping

Handler lists retain registration order and cache a sorted, immutable snapshot.
Registration invalidates the cache; a subsequent lookup rebuilds it under the
converter lock. Active render loops keep their original snapshot while later
lookups see new registrations. Rendering does not repeatedly copy and sort
handler lists for every node.

Escaping reads the character set under one read lock per text segment.
Unescaping loads its handler snapshot only on the first eligible marker in each
call, then reuses it for the remaining markers. Plain table cells therefore
avoid an unnecessary handler lookup.

## Development measurement

On Apple M5 / darwin-arm64 with Go 1.27.3, five 250 ms samples per workload
compared main revision `3cdb486` (using the same benchmark file) with the handler
snapshot and escaping changes. The following are medians for `Reuse`:

- Dense Markdown text: 1.469 ms to 0.674 ms; 2,904,369 to 791,733 bytes/op;
  32,117 to 94 allocations/op.
- Article sections: 0.781 ms to 0.657 ms; 16,198 to 10,569 allocations/op.
- Many sibling nodes: 1.254 ms to 1.109 ms; 33,282 to 20,990 allocations/op.
- Table: 0.472 ms to 0.430 ms; 12,655 to 8,603 allocations/op.

These are local development measurements, not release guarantees or results
from a representative production corpus. Both revisions used Go 1.27.3, so the
comparison measures the code changes rather than a Go toolchain upgrade. The
small-fragment `Fresh` case improved only about 3%; the dense punctuation case
must not be used to claim a twofold improvement for all HTML documents.
