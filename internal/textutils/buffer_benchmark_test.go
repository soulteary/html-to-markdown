package textutils

import (
	"bytes"
	"testing"
)

var bufferBenchmarkResult []byte

func BenchmarkNewlineBuffer(b *testing.B) {
	for _, tc := range []struct {
		name  string
		input []byte
	}{
		{"MostlyNewlines", append(append([]byte("start"), bytes.Repeat([]byte("\n "), 32768)...), []byte("end")...)},
		{"LongPlain", bytes.Repeat([]byte("plain text with spaces "), 2048)},
		{"LongUnicode", bytes.Repeat([]byte("中文 🌟 text  \n\n\n"), 2048)},
		{"LongSpaceRuns", bytes.Repeat([]byte("before"+string(bytes.Repeat([]byte(" "), 512))+"\n\n\nafter"), 64)},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tc.input)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bufferBenchmarkResult = TrimConsecutiveNewlines(tc.input)
			}
		})
	}
}
