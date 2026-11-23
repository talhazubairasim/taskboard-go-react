// Package metrics records request and query latencies so the numbers quoted
// in the README can be checked on a running server at GET /metrics.
package metrics

import (
	"context"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

// Recorder keeps the most recent samples in a ring buffer.
type Recorder struct {
	mu      sync.Mutex
	samples []time.Duration
	next    int
	full    bool
	count   int64
}

// NewRecorder keeps up to size samples.
func NewRecorder(size int) *Recorder {
	return &Recorder{samples: make([]time.Duration, size)}
}

// Observe records one duration.
func (r *Recorder) Observe(d time.Duration) {
	r.mu.Lock()
	r.samples[r.next] = d
	r.next = (r.next + 1) % len(r.samples)
	if r.next == 0 {
		r.full = true
	}
	r.count++
	r.mu.Unlock()
}

// Count is the total number of observations ever made.
func (r *Recorder) Count() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.count
}

// Percentile returns the p-th percentile (0-100) of the retained samples.
func (r *Recorder) Percentile(p float64) time.Duration {
	r.mu.Lock()
	n := r.next
	if r.full {
		n = len(r.samples)
	}
	s := make([]time.Duration, n)
	copy(s, r.samples[:n])
	r.mu.Unlock()
	if n == 0 {
		return 0
	}
	sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
	idx := int(float64(n)*p/100+0.5) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= n {
		idx = n - 1
	}
	return s[idx]
}

// Middleware times every HTTP request except WebSocket upgrades.
func (r *Recorder) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Upgrade") != "" {
			next.ServeHTTP(w, req)
			return
		}
		start := time.Now()
		next.ServeHTTP(w, req)
		r.Observe(time.Since(start))
	})
}

type startKey struct{}

// QueryTracer is a pgx tracer that records how long each SQL query takes.
type QueryTracer struct{ R *Recorder }

// TraceQueryStart implements pgx.QueryTracer.
func (t QueryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	return context.WithValue(ctx, startKey{}, time.Now())
}

// TraceQueryEnd implements pgx.QueryTracer.
func (t QueryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryEndData) {
	if start, ok := ctx.Value(startKey{}).(time.Time); ok {
		t.R.Observe(time.Since(start))
	}
}
