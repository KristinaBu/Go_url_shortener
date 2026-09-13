FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /url-shortener \
    ./cmd/server

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /url-shortener /app/url-shortener

EXPOSE 8080

ENTRYPOINT ["/app/url-shortener"]