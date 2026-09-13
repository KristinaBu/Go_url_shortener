run:
	go run ./cmd/server -storage memory

run-postgres:
	docker compose up -d

test:
	go test ./...

race:
	go test -race ./...

bench:
	go test -run=^$ -bench=BenchmarkHTTPGet -benchmem ./internal/bench

load-test:
	go test -v ./internal/bench -run TestHTTPGetLoad