FROM golang:1.16-buster AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/money-service cmd/money/main.go
FROM alpine:3.11.5
RUN adduser -D app-executor
USER app-executor
WORKDIR /app
COPY --from=builder /app/bin/money-service .
COPY --from=builder /app/data/postgres/migrations/money migrations
CMD ["./money-service"]