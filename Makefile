run-memory:
	docker compose -f docker-compose.memory.yml up

run-postgres:
	docker compose up -d

test:
	go test ./...

race:
	go test -race ./...

bench:
	go test -run=^$$ -bench=BenchmarkHTTPGet -benchmem ./internal/handler

load-test:
	go test -v ./internal/handler -run TestHTTPGetLoad