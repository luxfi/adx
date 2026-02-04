// Package examples provides verification tests for optimizations
package examples

import (
	"testing"
	"time"

	"github.com/luxfi/cache"
	"github.com/luxfi/fasthttp"
	"github.com/luxfi/metric"
	"github.com/luxfi/pool"
	"github.com/prebid/openrtb/v20/openrtb2"
)

// TestOptimizedComponents verifies all optimization components work together
func TestOptimizedComponents(t *testing.T) {
	t.Run("Memory Pooling", func(t *testing.T) {
		// Test byte slice pooling
		buf := pool.GetByteSlice()
		if buf == nil {
			t.Error("Failed to get byte slice from pool")
		}
		if cap(buf) == 0 {
			t.Error("Byte slice has zero capacity")
		}
		pool.PutByteSlice(buf)

		// Test string pooling
		s := pool.GetString()
		if s == "" {
			t.Error("Failed to get string from pool")
		}
		pool.PutString(s)

		// Test interface slice pooling
		is := pool.GetInterfaceSlice()
		if is == nil {
			t.Error("Failed to get interface slice from pool")
		}
		pool.PutInterfaceSlice(is)

		// Test map pooling
		m := pool.GetMap()
		if m == nil {
			t.Error("Failed to get map from pool")
		}
		pool.PutMap(m)

		// Test FastBuffer
		fb := pool.GetFastBuffer()
		if fb == nil {
			t.Error("Failed to get FastBuffer from pool")
		}
		fb.Write([]byte("test"))
		if fb.String() != "test" {
			t.Error("FastBuffer string conversion failed")
		}
		pool.PutFastBuffer(fb)
	})

	t.Run("Optimized Metrics", func(t *testing.T) {
		// Test metrics registry
		reg := metric.NewRegistry()
		if reg == nil {
			t.Error("Failed to create metrics registry")
		}

		// Test counter
		counter := reg.NewCounter("test_counter", "Test counter")
		if counter == nil {
			t.Error("Failed to create counter")
		}

		counter.Inc()
		counter.Inc()
		counter.Add(3.5)

		if counter.Get() != 5.5 {
			t.Errorf("Counter value mismatch: got %f, want 5.5", counter.Get())
		}

		// Test gauge
		gauge := reg.NewGauge("test_gauge", "Test gauge")
		if gauge == nil {
			t.Error("Failed to create gauge")
		}

		gauge.Set(10.5)
		gauge.Inc()
		gauge.Dec()
		gauge.Add(2.5)
		gauge.Sub(1.0)

		if gauge.Get() != 12.0 {
			t.Errorf("Gauge value mismatch: got %f, want 12.0", gauge.Get())
		}

		// Test histogram
		histogram := reg.NewHistogram("test_histogram", "Test histogram", []float64{1.0, 2.0, 3.0})
		if histogram == nil {
			t.Error("Failed to create histogram")
		}

		histogram.Observe(1.5)
		histogram.Observe(2.5)
		histogram.Observe(3.5)

		// Test timing metric
		hist := reg.NewHistogram("test_timing", "Test timing", []float64{0.001, 0.01, 0.1, 1.0})
		timer := metric.NewTimingMetric(hist)

		time.Sleep(10 * time.Millisecond)
		timer.Stop()

		if timer.Duration() < 10*time.Millisecond {
			t.Error("Timer duration too short")
		}
	})

	t.Run("FastHTTP Server", func(t *testing.T) {
		// Test FastHTTP server creation
		handler := &testHandler{}
		server := fasthttp.NewServer(handler)

		if server == nil {
			t.Error("Failed to create FastHTTP server")
		}

		// Test router
		router := fasthttp.NewRouter()
		if router == nil {
			t.Error("Failed to create router")
		}

		// Add a test route
		router.AddRoute("GET", "/test", func(ctx *fasthttp.RequestCtx) {
			ctx.WriteString("OK")
		})

		// Test middleware
		middleware := func(next fasthttp.RouteHandler) fasthttp.RouteHandler {
			return func(ctx *fasthttp.RequestCtx) {
				ctx.WriteString("Middleware: ")
				next(ctx)
			}
		}

		chained := fasthttp.ChainMiddleware(func(ctx *fasthttp.RequestCtx) {
			ctx.WriteString("Handler")
		}, middleware)

		if chained == nil {
			t.Error("Failed to chain middleware")
		}
	})

	t.Run("Cache Integration", func(t *testing.T) {
		cache := cache.NewLRU[string, openrtb2.BidRequest](100)

		if cache == nil {
			t.Error("Failed to create cache")
		}

		// Test cache operations
		request := openrtb2.BidRequest{
			ID: "test-request",
			Imp: []openrtb2.Imp{
				{
					ID: "imp-1",
					Banner: &openrtb2.Banner{
						W: 300,
						H: 250,
					},
				},
			},
		}

		// Put and get
		cache.Put("key1", request)
		val, ok := cache.Get("key1")
		if !ok {
			t.Error("Failed to get value from cache")
		}
		if val.ID != "test-request" {
			t.Error("Cache returned wrong value")
		}

		// Test miss
		_, ok = cache.Get("nonexistent")
		if ok {
			t.Error("Cache should not have returned value for nonexistent key")
		}

		// Test size
		if cache.Size() != 1 {
			t.Errorf("Cache size mismatch: got %d, want 1", cache.Size())
		}

		// Test eviction
		for i := 0; i < 100; i++ {
			cache.Put(fmt.Sprintf("key-%d", i), request)
		}

		if cache.Size() != 100 {
			t.Errorf("Cache size after eviction: got %d, want 100", cache.Size())
		}
	})
}

// testHandler is a simple handler for testing
type testHandler struct{}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// TestPerformanceComparison compares standard vs optimized implementations
func TestPerformanceComparison(t *testing.T) {
	// Test memory allocation performance
	start := time.Now()
	for i := 0; i < 10000; i++ {
		buf := make([]byte, 1024)
		_ = buf
	}
	standardTime := time.Since(start)

	start = time.Now()
	for i := 0; i < 10000; i++ {
		buf := pool.GetByteSlice()
		pool.PutByteSlice(buf)
	}
	optimizedTime := time.Since(start)

	t.Logf("Memory allocation performance:")
	t.Logf("  Standard: %v", standardTime)
	t.Logf("  Optimized: %v", optimizedTime)
	t.Logf("  Improvement: %.1fx faster", float64(standardTime)/float64(optimizedTime))

	// Verify improvement
	if float64(standardTime)/float64(optimizedTime) < 1.5 {
		t.Error("Optimized allocation should be at least 1.5x faster")
	}
}

// TestIntegrationExample verifies the complete integration example works
func TestIntegrationExample(t *testing.T) {
	// Create a simple optimized server
	reg := metric.NewRegistry()
	cache := cache.NewLRU[string, openrtb2.BidRequest](100)
	handler := &BidderHandler{
		metrics: reg,
		cache:   cache,
	}

	server := fasthttp.NewServer(handler)
	if server == nil {
		t.Error("Failed to create integrated server")
	}

	// Verify metrics are working
	counter := reg.GetCounter("requests")
	if counter == nil {
		t.Error("Metrics not properly integrated")
	}

	// Verify cache is working
	request := openrtb2.BidRequest{ID: "test"}
	cache.Put("test-key", request)

	val, ok := cache.Get("test-key")
	if !ok || val.ID != "test" {
		t.Error("Cache integration failed")
	}

	t.Log("✅ Complete integration example works correctly")
}

// Example usage:
// go test -v ./examples/
// go test -bench=. -benchmem ./examples/
