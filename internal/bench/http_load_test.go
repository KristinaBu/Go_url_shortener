package bench

import (
	"fmt"
	"net/http"
	"sort"
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
		wg         sync.WaitGroup
		success    atomic.Int64
		errors     atomic.Int64
		errorMu    sync.Mutex
		firstError string
	)

	latencies := make([]time.Duration, 0, totalRequests)
	var latencyMu sync.Mutex

	baseRequests := totalRequests / concurrency
	extraRequests := totalRequests % concurrency

	start := time.Now()

	for worker := 0; worker < concurrency; worker++ {
		requestsForWorker := baseRequests

		if worker < extraRequests {
			requestsForWorker++
		}

		wg.Add(1)

		go func() {
			defer wg.Done()

			for i := 0; i < requestsForWorker; i++ {
				requestStart := time.Now()

				req, err := http.NewRequest(
					http.MethodGet,
					benchURL,
					nil,
				)
				if err != nil {
					errors.Add(1)

					errorMu.Lock()
					if firstError == "" {
						firstError = err.Error()
					}
					errorMu.Unlock()

					latency := time.Since(requestStart)

					latencyMu.Lock()
					latencies = append(latencies, latency)
					latencyMu.Unlock()

					continue
				}

				resp, err := client.Do(req)
				latency := time.Since(requestStart)

				if err != nil {
					errors.Add(1)

					errorMu.Lock()
					if firstError == "" {
						firstError = err.Error()
					}
					errorMu.Unlock()

					latencyMu.Lock()
					latencies = append(latencies, latency)
					latencyMu.Unlock()

					continue
				}

				resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					errors.Add(1)

					errorMu.Lock()
					if firstError == "" {
						firstError = fmt.Sprintf(
							"unexpected HTTP status: %d",
							resp.StatusCode,
						)
					}
					errorMu.Unlock()
				} else {
					success.Add(1)
				}

				latencyMu.Lock()
				latencies = append(latencies, latency)
				latencyMu.Unlock()
			}
		}()
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

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	var totalLatency time.Duration

	for _, latency := range latencies {
		totalLatency += latency
	}

	avgLatency := totalLatency / time.Duration(len(latencies))

	p95Index := int(float64(len(latencies))*0.95) - 1
	if p95Index < 0 {
		p95Index = 0
	}
	if p95Index >= len(latencies) {
		p95Index = len(latencies) - 1
	}

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
