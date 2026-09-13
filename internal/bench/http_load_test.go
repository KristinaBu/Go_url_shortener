package bench

import (
	"fmt"
	"net/http"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	benchURL       = "http://localhost:8080/links/NSILa0pd7S"
	requestTimeout = 5 * time.Second
)

func BenchmarkHTTPGet(b *testing.B) {
	client := &http.Client{
		Timeout: requestTimeout,
	}

	b.ResetTimer()

	for range b.N {
		resp, err := client.Get(benchURL)
		if err != nil {
			b.Fatal(err)
		}

		resp.Body.Close()
	}
}

func runLoadTest(t *testing.T, concurrency int, totalRequests int) {
	t.Helper()

	transport := &http.Transport{
		MaxIdleConns:        concurrency,
		MaxIdleConnsPerHost: concurrency,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   requestTimeout,
	}

	defer transport.CloseIdleConnections()

	var (
		wg      sync.WaitGroup
		success atomic.Int64
		errors  atomic.Int64
	)

	latencies := make([]time.Duration, 0, totalRequests)

	var (
		latencyMu  sync.Mutex
		errorMu    sync.Mutex
		firstError string
	)

	baseRequests := totalRequests / concurrency
	extraRequests := totalRequests % concurrency

	start := time.Now()

	for worker := range concurrency {
		requestsForWorker := baseRequests

		if worker < extraRequests {
			requestsForWorker++
		}

		wg.Go(func() {
			for range requestsForWorker {
				requestStart := time.Now()

				req, err := http.NewRequest(
					http.MethodGet,
					benchURL,
					nil,
				)
				if err != nil {
					errors.Add(1)
					recordError(&errorMu, &firstError, err.Error())
					recordLatency(&latencyMu, &latencies, time.Since(requestStart))
					continue
				}

				resp, err := client.Do(req)
				latency := time.Since(requestStart)

				if err != nil {
					errors.Add(1)
					recordError(&errorMu, &firstError, err.Error())
					recordLatency(&latencyMu, &latencies, latency)
					continue
				}

				resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					errors.Add(1)
					recordError(
						&errorMu,
						&firstError,
						fmt.Sprintf("unexpected HTTP status: %d", resp.StatusCode),
					)
				} else {
					success.Add(1)
				}

				recordLatency(&latencyMu, &latencies, latency)
			}
		})
	}

	wg.Wait()

	duration := time.Since(start)

	if len(latencies) != totalRequests {
		t.Fatalf(
			"expected %d requests, got %d",
			totalRequests,
			len(latencies),
		)
	}

	slices.Sort(latencies)

	var totalLatency time.Duration

	for _, latency := range latencies {
		totalLatency += latency
	}

	avgLatency := totalLatency / time.Duration(len(latencies))

	p95Index := max(int(float64(len(latencies))*0.95)-1, 0)
	p95Latency := latencies[p95Index]

	successful := success.Load()
	failed := errors.Load()

	rps := float64(totalRequests) / duration.Seconds()
	successRPS := float64(successful) / duration.Seconds()

	errorMu.Lock()
	errText := firstError
	errorMu.Unlock()

	fmt.Printf(
		"\n"+
			"Concurrency: %d\n"+
			"Requests:    %d\n"+
			"Successful:  %d\n"+
			"Errors:      %d\n"+
			"RPS:         %.2f\n"+
			"Success RPS: %.2f\n"+
			"Avg latency: %s\n"+
			"P95 latency: %s\n"+
			"First error: %s\n",
		concurrency,
		totalRequests,
		successful,
		failed,
		rps,
		successRPS,
		avgLatency,
		p95Latency,
		errText,
	)
}

func recordLatency(
	mu *sync.Mutex,
	latencies *[]time.Duration,
	latency time.Duration,
) {
	mu.Lock()
	*latencies = append(*latencies, latency)
	mu.Unlock()
}

func recordError(
	mu *sync.Mutex,
	firstError *string,
	err string,
) {
	mu.Lock()
	if *firstError == "" {
		*firstError = err
	}
	mu.Unlock()
}

func TestHTTPGetLoad(t *testing.T) {
	resp, err := http.Get(benchURL)
	if err != nil {
		t.Skipf("server is not running: %v", err)
	}
	resp.Body.Close()

	for _, concurrency := range []int{10, 50, 100} {
		t.Run(
			fmt.Sprintf("concurrency_%d", concurrency),
			func(t *testing.T) {
				totalRequests := concurrency
				runLoadTest(t, concurrency, totalRequests)
			},
		)
	}
}
