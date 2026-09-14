package handler

import (
	"fmt"
	"github.com/KristinaBu/Go_url_shortener/internal/domain"
	"github.com/KristinaBu/Go_url_shortener/pkg/cache"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/KristinaBu/Go_url_shortener/internal/repository"
	"github.com/KristinaBu/Go_url_shortener/internal/service"
	"github.com/KristinaBu/Go_url_shortener/pkg/generator"
)

const requestTimeout = 5 * time.Second

func newTestServer(tb testing.TB) (*httptest.Server, string) {
	tb.Helper()

	repo := repository.NewMemoryRepository()

	linkCache, err := cache.NewLRU[string, domain.Link](1000)
	if err != nil {
		tb.Fatal(err)
	}

	gen := generator.New()
	svc := service.NewLinkService(repo, gen, linkCache)
	h := New(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("/links", h.CreateLink)
	mux.HandleFunc("/links/", h.GetLink)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	httpHandler := LoggingMiddleware(
		logger,
		mux,
	)

	server := httptest.NewServer(httpHandler)

	link, err := svc.Create(
		tb.Context(),
		"https://example.com",
	)
	if err != nil {
		server.Close()
		tb.Fatal(err)
	}

	return server, server.URL + "/links/" + link.ShortCode
}

func BenchmarkHTTPGet(b *testing.B) {
	server, benchURL := newTestServer(b)
	defer server.Close()

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

func runLoadTest(
	t *testing.T,
	client *http.Client,
	benchURL string,
	concurrency int,
	totalRequests int,
) {
	t.Helper()

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

				resp, err := client.Get(benchURL)
				latency := time.Since(requestStart)

				if err != nil {
					errors.Add(1)
					recordError(&errorMu, &firstError, err.Error())
					recordLatency(&latencyMu, &latencies, latency)
					continue
				}

				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					errors.Add(1)
					recordError(
						&errorMu,
						&firstError,
						fmt.Sprintf(
							"unexpected HTTP status: %d",
							resp.StatusCode,
						),
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

	p95Index := int(float64(len(latencies))*0.95) - 1
	p95Index = max(p95Index, 0)

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
		latencies[p95Index],
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
	server, benchURL := newTestServer(t)
	defer server.Close()

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		MaxConnsPerHost:     100,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   requestTimeout,
	}

	defer transport.CloseIdleConnections()

	const totalRequests = 3_000

	for _, concurrency := range []int{10, 50, 100} {
		t.Run(
			fmt.Sprintf("concurrency_%d", concurrency),
			func(t *testing.T) {
				runLoadTest(
					t,
					client,
					benchURL,
					concurrency,
					totalRequests,
				)
			},
		)
	}
}
