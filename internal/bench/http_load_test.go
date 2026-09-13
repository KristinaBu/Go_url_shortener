package bench

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	benchURL       = "http://localhost:8080/links/NSILa0pd7S"
	requestsPerRun = 10000
	requestTimeout = 5 * time.Second
)

func BenchmarkHTTPGet(b *testing.B) {
	client := &http.Client{
		Timeout: requestTimeout,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			benchURL,
			nil,
		)
		if err != nil {
			b.Fatal(err)
		}

		resp, err := client.Do(req)
		if err != nil {
			b.Fatal(err)
		}

		resp.Body.Close()
	}
}

func runLoadTest(
	t *testing.T,
	concurrency int,
	totalRequests int,
) {
	t.Helper()

	client := &http.Client{
		Timeout: requestTimeout,
	}

	var (
		wg       sync.WaitGroup
		requests atomic.Int64
		errors   atomic.Int64
	)

	latencies := make([]time.Duration, 0, totalRequests)
	var latencyMu sync.Mutex

	start := time.Now()

	for worker := 0; worker < concurrency; worker++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for {
				n := int(requests.Add(1))
				if n > totalRequests {
					return
				}

				requestStart := time.Now()

				req, err := http.NewRequest(
					http.MethodGet,
					benchURL,
					nil,
				)
				if err != nil {
					errors.Add(1)
					continue
				}

				resp, err := client.Do(req)
				if err != nil {
					errors.Add(1)
					continue
				}

				resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					errors.Add(1)
				}

				latency := time.Since(requestStart)

				latencyMu.Lock()
				latencies = append(latencies, latency)
				latencyMu.Unlock()
			}
		}()
	}

	wg.Wait()

	duration := time.Since(start)

	if len(latencies) == 0 {
		t.Fatal("no successful requests")
	}

	// Сортируем задержки.
	for i := 0; i < len(latencies); i++ {
		for j := i + 1; j < len(latencies); j++ {
			if latencies[j] < latencies[i] {
				latencies[i], latencies[j] = latencies[j], latencies[i]
			}
		}
	}

	var totalLatency time.Duration
	for _, latency := range latencies {
		totalLatency += latency
	}

	avgLatency := totalLatency / time.Duration(len(latencies))

	p95Index := int(float64(len(latencies)) * 0.95)
	if p95Index >= len(latencies) {
		p95Index = len(latencies) - 1
	}

	p95Latency := latencies[p95Index]

	rps := float64(totalRequests) / duration.Seconds()

	fmt.Printf(
		"\nConcurrency: %d\n"+
			"Requests:    %d\n"+
			"RPS:         %.2f\n"+
			"Avg latency: %s\n"+
			"P95 latency: %s\n"+
			"Errors:      %d\n",
		concurrency,
		totalRequests,
		rps,
		avgLatency,
		p95Latency,
		errors.Load(),
	)
}

func TestHTTPGetLoad(t *testing.T) {
	if _, err := http.Get(benchURL); err != nil {
		t.Skip("server is not running")
	}

	for _, concurrency := range []int{10, 50, 100, 500} {
		t.Run(fmt.Sprintf("concurrency_%d", concurrency), func(t *testing.T) {
			runLoadTest(t, concurrency, requestsPerRun)
		})
	}
}
