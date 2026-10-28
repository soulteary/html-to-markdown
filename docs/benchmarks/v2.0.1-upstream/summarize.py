#!/usr/bin/env python3
"""Summarize one retained Go benchmark run without third-party dependencies."""
import argparse
import csv
import json
from pathlib import Path
import random
import re
import statistics

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('raw', type=Path)
parser.add_argument('--csv', type=Path, required=True)
parser.add_argument('--json', type=Path)
args = parser.parse_args()
raw = args.raw.read_text(encoding='utf-8')
if not re.search(r'^PASS$', raw, re.M) or re.search(r'^FAIL(?:\s|$)', raw, re.M):
    parser.error('benchmark output does not contain a successful PASS result')
pattern = re.compile(
    r'BenchmarkV201VsUpstream/(\w+)/(Fresh|Reuse)/Round(\d+)/'
    r'(Upstream|Fork)(?:-\d+)?\s+\d+\s+([\d.]+) ns/op.*?'
    r'([\d.]+) B/op\s+([\d.]+) allocs/op'
)
cases = {}
for match in pattern.finditer(raw):
    case, mode, round_id, version, ns, byte_count, allocations = match.groups()
    sample = cases.setdefault((case, mode), {}).setdefault(int(round_id), {})
    if version in sample:
        parser.error('duplicate sample; expected a single -count=1 run')
    sample[version] = [float(ns), float(byte_count), float(allocations)]
if len(cases) != 18:
    parser.error('expected nine workloads with Fresh and Reuse modes')
random.seed(20261028)
rows = []
for (case, mode), rounds in cases.items():
    if set(rounds) != set(range(8)) or any(set(p) != {'Upstream', 'Fork'} for p in rounds.values()):
        parser.error('expected eight complete paired rounds per workload/mode')
    pairs = [rounds[i] for i in range(8)]
    upstream = [statistics.median(p['Upstream'][i] for p in pairs) for i in range(3)]
    fork = [statistics.median(p['Fork'][i] for p in pairs) for i in range(3)]
    delta = [(fork[i] / upstream[i] - 1) * 100 for i in range(3)]
    bootstrap = []
    for _ in range(5000):
        picked = random.choices(pairs, k=8)
        bootstrap.append((statistics.median(p['Fork'][0] for p in picked) /
                          statistics.median(p['Upstream'][0] for p in picked) - 1) * 100)
    bootstrap.sort()
    rows.append(dict(case=case, mode=mode, upstream_ns=upstream[0], fork_ns=fork[0],
                     time_change_pct=delta[0], speedup=upstream[0] / fork[0],
                     upstream_bytes=upstream[1], fork_bytes=fork[1], bytes_change_pct=delta[1],
                     upstream_allocs=upstream[2], fork_allocs=fork[2], allocs_change_pct=delta[2],
                     time_change_bootstrap_95=[bootstrap[125], bootstrap[4874]]))
fields = [key for key in rows[0] if key != 'time_change_bootstrap_95']
with args.csv.open('w', encoding='utf-8', newline='') as file:
    writer = csv.DictWriter(file, fieldnames=fields, lineterminator="\n")
    writer.writeheader()
    writer.writerows({key: row[key] for key in fields} for row in rows)
if args.json:
    args.json.write_text(json.dumps(rows, indent=2), encoding='utf-8')
print('Summarized 18 workload/mode combinations from 288 timed samples.')
