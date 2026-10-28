# v2.0.1 vs upstream v2.5.2 benchmark

Both dependencies are pinned to public module releases without replacements.
Upstream v2.5.2 corresponds to 290df46a279e3d7d9011dbbb199658dfcb5ec272.
Fork v2.0.1 corresponds to 1ddc200789715a7322354988ea3b52f64e5d2948.
The common HTML parser dependency is golang.org/x/net v0.55.0.

From this directory, with network access for initial module downloads:

```sh
GOWORK=off GOTOOLCHAIN=go1.27.3 GOMAXPROCS=1 go test -run '^TestInputOutputParity$' -v
GOWORK=off GOTOOLCHAIN=go1.27.3 GOMAXPROCS=1 go test -run '^$' -bench BenchmarkV201VsUpstream -benchmem -benchtime=500ms -count=1 > raw.txt
```

The benchmark itself contains eight paired rounds; do not set `-count=8` unless
additional rounds are intended. It alternates which version runs first (four
rounds each), with no parallel benchmark execution. Fresh recreates a converter
per operation (table tests include the table plugin); Reuse configures it once
outside the timed loop. Every operation still parses a new HTML DOM. All input
generators match the v2.0.1 end-to-end benchmark suite. Output equality is checked
before timing; the parity test additionally prints input/output SHA-256 hashes.

Timing includes HTML parsing, normalization, rendering, escaping, and the result.
Module loading, compilation, and converter construction in Reuse are outside the
timed loop. GC costs during the loop remain included. `B/op` means allocated
bytes per operation, not peak/live memory or RSS. `allocs/op` is allocation count.

The report uses medians of eight samples for ns/op, B/op, and allocs/op. Change
percentage = (fork median / upstream median - 1) * 100. Approximate timing ranges
use 5,000 paired bootstrap resamples with seed 20261028, resampling the eight
rounds and recomputing the ratio of medians. They describe this sample only,
not uncertainty from all environments or a production guarantee. The primary
run uses GOMAXPROCS=1; the retained default-settings control run used 10. Results
from the two settings are never combined. This is not a parallel throughput test.

## Retained measurements

The primary `results/gomaxprocs-1.txt` and supplemental
`results/gomaxprocs-10.txt` contain the original successful benchmark output.
Their `.csv` companions use raw ns/op and bytes/op units. Input/output hashes
and the toolchain record are also in `results/`. The main results and input
workloads are explained in [PERFORMANCE.md](../../../PERFORMANCE.md#v201-compared-with-upstream-v252).

Python 3 (standard library only) can reproduce the summaries without rerunning
the measurements:

```sh
python3 summarize.py results/gomaxprocs-1.txt --csv summary.csv --json summary.json
python3 summarize.py results/gomaxprocs-10.txt --csv control.csv --json control.json
```

This directory is a standalone measurement module, not part of the library's
runtime dependency graph. Its tests import the published fork version as well
as upstream; they do not benchmark the current checkout implicitly. Keep
`GOWORK=off` and the pinned module versions when reproducing this comparison.
