package proxmox

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_logCache_Load(t *testing.T) {
	t.Parallel()
	array := func(size int) []string {
		out := make([]string, size)
		for i := range size {
			out[i] = strconv.Itoa(i)
		}
		return out
	}

	newCache := func(index uint64, chunk *logChunk) *logCache {
		var aa logCache
		aa.numberOfElements.Store(index)
		aa.Chunks = chunk
		return &aa
	}

	newChunk := func(start, size int) [chunkSize]string {
		var chunk [chunkSize]string
		for i := range size {
			chunk[i] = strconv.Itoa(start + i)
		}
		return chunk
	}

	tests := []struct {
		name   string
		input  *logCache
		output []string
	}{
		{name: `no item`,
			input: newCache(0, &logChunk{
				chunk: newChunk(0, 0)}),
			output: array(0)},
		{name: `single item`,
			input: newCache(1, &logChunk{
				chunk: newChunk(0, 1)}),
			output: array(1)},
		{name: `single full chunk`,
			input: newCache(chunkSize, &logChunk{
				chunk: newChunk(0, chunkSize)}),
			output: array(chunkSize)},
		{name: `single full chunk +1`,
			input: newCache(chunkSize+1, &logChunk{
				chunk: newChunk(0, chunkSize),
				next: &logChunk{
					chunk: newChunk(chunkSize, 1)}}),
			output: array(chunkSize + 1)},
		{name: `single full chunk -1`,
			input: newCache(chunkSize-1, &logChunk{
				chunk: newChunk(0, chunkSize-1)}),
			output: array(chunkSize - 1)},
		{name: `multiple full chunk`,
			input: newCache(chunkSize*3, &logChunk{
				chunk: newChunk(0, chunkSize),
				next: &logChunk{
					chunk: newChunk(chunkSize*1, chunkSize),
					next: &logChunk{
						chunk: newChunk(chunkSize*2, chunkSize),
						next: &logChunk{
							chunk: newChunk(chunkSize*3, chunkSize),
							next: &logChunk{
								chunk: newChunk(chunkSize*4, chunkSize)}}}}}),
			output: array(chunkSize * 3)},
		{name: `multiple full chunk +1`,
			input: newCache(chunkSize*3+1, &logChunk{
				chunk: newChunk(0, chunkSize),
				next: &logChunk{
					chunk: newChunk(chunkSize*1, chunkSize),
					next: &logChunk{
						chunk: newChunk(chunkSize*2, chunkSize),
						next: &logChunk{
							chunk: newChunk(chunkSize*3, chunkSize),
							next: &logChunk{
								chunk: newChunk(chunkSize*4, chunkSize)}}}}}),
			output: array(chunkSize*3 + 1)},
		{name: `multiple full chunk -1`,
			input: newCache(chunkSize*3-1, &logChunk{
				chunk: newChunk(0, chunkSize),
				next: &logChunk{
					chunk: newChunk(chunkSize*1, chunkSize),
					next: &logChunk{
						chunk: newChunk(chunkSize*2, chunkSize),
						next: &logChunk{
							chunk: newChunk(chunkSize*3, chunkSize),
							next: &logChunk{
								chunk: newChunk(chunkSize*4, chunkSize)}}}}}),
			output: array(chunkSize*3 - 1)},
		{name: `multiple chunk`,
			input: newCache(chunkSize*2+32, &logChunk{
				chunk: newChunk(0, chunkSize),
				next: &logChunk{
					chunk: newChunk(chunkSize*1, chunkSize),
					next: &logChunk{
						chunk: newChunk(chunkSize*2, chunkSize),
						next: &logChunk{
							chunk: newChunk(chunkSize*2+64, chunkSize)}}}}),
			output: array(chunkSize*2 + 32)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			aaa := test.input.Load()
			require.Len(t, aaa, len(test.output))
			require.Equal(t, test.output, aaa)
		})
	}
}

func Benchmark_logCache_Load(b *testing.B) {
	arrayMap := func(size int) []any {
		out := make([]any, size)
		for i := range out {
			out[i] = map[string]any{
				"t": strconv.Itoa(i),
			}
		}
		return out
	}

	for _, size := range []int{
		1, 255, 256, 257,
		1024, 4096, 16384, 65536,
	} {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			input := arrayMap(size)

			cache := &logCache{}
			chunk := &logChunk{}
			cache.Chunks = chunk
			cache.Current = chunk
			cache.Store(input)

			var sink []string

			b.ResetTimer()
			for b.Loop() {
				sink = cache.Load()
			}

			_ = sink
		})
	}
}

func Test_logCache_Store(t *testing.T) {
	t.Parallel()

	arrayMap := func(start, size int) []any {
		out := make([]any, size)
		for i := range len(out) {
			out[i] = map[string]any{
				"t": strconv.Itoa(start + i),
			}
		}
		return out
	}

	array := func(size int) []string {
		out := make([]string, size)
		for i := range size {
			out[i] = strconv.Itoa(i)
		}
		return out
	}

	tests := []struct {
		name   string
		input  [][]any
		output []string
	}{
		{name: `single item`,
			input:  [][]any{arrayMap(0, 1)},
			output: array(1)},
		{name: `single full chunk`,
			input:  [][]any{arrayMap(0, chunkSize)},
			output: array(chunkSize)},
		{name: `single full chunk +1`,
			input:  [][]any{arrayMap(0, chunkSize+1)},
			output: array(chunkSize + 1)},
		{name: `single full chunk -1`,
			input:  [][]any{arrayMap(0, chunkSize-1)},
			output: array(chunkSize - 1)},
		{name: `single half chunk`,
			input:  [][]any{arrayMap(0, 128)},
			output: array(128)},
		{name: `multiple full chunk`,
			input: [][]any{
				arrayMap(0, chunkSize*1),
				arrayMap(chunkSize*1, chunkSize*1),
				arrayMap(chunkSize*2, chunkSize*1)},
			output: array(chunkSize * 3)},
		{name: `multiple full chunk +1`,
			input: [][]any{
				arrayMap(0, chunkSize*1),
				arrayMap(chunkSize*1, chunkSize*1),
				arrayMap(chunkSize*2, chunkSize*1),
				arrayMap(chunkSize*3, 1)},
			output: array(chunkSize*3 + 1)},
		{name: `multiple full chunk -1`,
			input: [][]any{
				arrayMap(0, chunkSize*1),
				arrayMap(chunkSize*1, chunkSize*1),
				arrayMap(chunkSize*2, chunkSize*1-1)},
			output: array(chunkSize*3 - 1)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cache := &logCache{}
			newChunk := &logChunk{}
			cache.Chunks = newChunk
			cache.Current = newChunk
			for _, e := range test.input {
				cache.Store(e)
			}
			output := cache.Load()
			require.Len(t, output, len(test.output))
			require.Equal(t, test.output, output)
		})
	}
}

func Benchmark_logCache_Store(b *testing.B) {
	arrayMap := func(size int) []any {
		out := make([]any, size)
		for i := range len(out) {
			out[i] = map[string]any{
				"t": strconv.Itoa(i),
			}
		}
		return out
	}

	inputs := [][]any{
		arrayMap(1),
		arrayMap(255),
		arrayMap(256),
		arrayMap(257),
		arrayMap(1024),
		arrayMap(4096),
		arrayMap(16384),
	}
	for b.Loop() {
		for i := range inputs {
			cache := &logCache{}
			newChunk := &logChunk{}
			cache.Chunks = newChunk
			cache.Current = newChunk
			cache.Store(inputs[i])
		}
	}
}
