package httputil

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

type releaseCapacityReader struct{}

func (releaseCapacityReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 'x'
	}
	return len(p), nil
}

// Opt in explicitly in an isolated memory-limited container, not on every CI job.
func TestReleaseV0220LargeBodyBudgetLifecycle(t *testing.T) {
	if os.Getenv("SUB2API_RELEASE_CAPACITY") != "1" {
		t.Skip("set SUB2API_RELEASE_CAPACITY=1 for the large-body capacity experiment")
	}
	for _, limitMiB := range []int64{256, 512, 1024} {
		t.Run(fmt.Sprintf("budget_%d_MiB", limitMiB), func(t *testing.T) {
			budget := NewBodyMemoryBudget(limitMiB << 20)
			type result struct {
				body    []byte
				err     error
				release func()
			}
			results := make(chan result, 8)
			start := make(chan struct{})
			var wg sync.WaitGroup
			for i := 0; i < cap(results); i++ {
				size := int64(59 << 20)
				if i%2 == 1 {
					size = 85 << 20
				}
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-start
					ctx, release := WithRequestBodyMemoryBudget(context.Background(), budget)
					req := httptest.NewRequest("POST", "/v1/responses", io.LimitReader(releaseCapacityReader{}, size)).WithContext(ctx)
					req.ContentLength = -1
					body, err := ReadRequestBodyWithPrealloc(req)
					if err == nil && int64(len(body)) != size {
						err = fmt.Errorf("truncated body: got %d want %d", len(body), size)
					}
					results <- result{body, err, release}
				}()
			}
			close(start)
			wg.Wait()
			close(results)
			accepted, rejected := 0, 0
			for r := range results {
				if r.err == nil {
					accepted++
				} else {
					require.True(t, errors.Is(r.err, ErrRequestBodyMemoryBudget), "%v", r.err)
					rejected++
				}
				r.release()
				r.release()
				runtime.KeepAlive(r.body)
			}
			require.Equal(t, 8, accepted+rejected)
			if limitMiB < 576 {
				require.Positive(t, rejected)
			}
			budget.mu.Lock()
			used := budget.used
			budget.mu.Unlock()
			require.Zero(t, used)
			ctx, release := WithRequestBodyMemoryBudget(context.Background(), budget)
			req := httptest.NewRequest("POST", "/v1/responses", io.LimitReader(releaseCapacityReader{}, 1024)).WithContext(ctx)
			body, err := ReadRequestBodyWithPrealloc(req)
			require.NoError(t, err)
			require.Len(t, body, 1024)
			release()
			runtime.GC()
			var memory runtime.MemStats
			runtime.ReadMemStats(&memory)
			peak, _ := os.ReadFile("/sys/fs/cgroup/memory.peak")
			t.Logf("budget=%d MiB accepted=%d rejected=%d released=%d heap_after_gc=%d cgroup_peak=%s", limitMiB, accepted, rejected, used, memory.HeapAlloc, peak)
		})
	}
}
