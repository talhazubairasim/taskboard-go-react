package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestPercentile(t *testing.T) {
	r := NewRecorder(200)
	if r.Percentile(95) != 0 {
		t.Fatal("empty recorder should report 0")
	}
	for i := 1; i <= 100; i++ {
		r.Observe(time.Duration(i) * time.Millisecond)
	}
	if got := r.Percentile(95); got != 95*time.Millisecond {
		t.Fatalf("p95 = %v", got)
	}
	if got := r.Percentile(100); got != 100*time.Millisecond {
		t.Fatalf("p100 = %v", got)
	}
	if got := r.Percentile(0); got != time.Millisecond {
		t.Fatalf("p0 = %v", got)
	}
	if r.Count() != 100 {
		t.Fatal("count")
	}
}

func TestRingBufferWraps(t *testing.T) {
	r := NewRecorder(10)
	for i := 0; i < 10; i++ {
		r.Observe(time.Second)
	}
	for i := 0; i < 10; i++ {
		r.Observe(time.Millisecond)
	}
	if r.Percentile(100) != time.Millisecond {
		t.Fatal("old samples should have been overwritten")
	}
	if r.Count() != 20 {
		t.Fatal("count should include overwritten samples")
	}
}

func TestMiddlewareSkipsWebSockets(t *testing.T) {
	r := NewRecorder(10)
	h := r.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Upgrade", "websocket")
	h.ServeHTTP(httptest.NewRecorder(), req)
	if r.Count() != 1 {
		t.Fatalf("count = %d", r.Count())
	}
}

func TestQueryTracer(t *testing.T) {
	r := NewRecorder(10)
	tr := QueryTracer{R: r}
	ctx := tr.TraceQueryStart(context.Background(), nil, pgx.TraceQueryStartData{})
	tr.TraceQueryEnd(ctx, nil, pgx.TraceQueryEndData{})
	tr.TraceQueryEnd(context.Background(), nil, pgx.TraceQueryEndData{}) // no start: ignored
	if r.Count() != 1 {
		t.Fatal("tracer did not record exactly one query")
	}
}
