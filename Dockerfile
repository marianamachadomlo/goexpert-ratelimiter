FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /rate-limiter ./cmd/server

FROM golang:1.22-alpine AS test

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["go", "test", "./...", "-v", "-count=1"]

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /rate-limiter /app/rate-limiter

EXPOSE 8080

CMD ["/app/rate-limiter"]
