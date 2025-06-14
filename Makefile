.PHONY: clean test lint bench

all: clean bin/money test lint check-arch

clean:
	rm -rf bin/*

download:
	go mod download

bin/%:
	export GOARCH=amd64
	export GOOS=linux
	export CGO_ENABLED=0
	go build -o ./bin/$(notdir $@) ./cmd/$(notdir $@)

test:
	go test ./...

build:
	go build -o money.exe cmd/money/main.go

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out

lint:
	golangci-lint run

check-arch:
	go-cleanarch

up:
	docker-compose up -d --build

down:
	docker-compose down

bench/%:
	ab -p data.json -T application/json -H "Idempotency-Key: e3f4a717-de9c-4d42-8fa7-151f0268c525" -c $(notdir $@) -n 2000 http://localhost:8000/api/v1/money/transfer

bench2/%:
	ab -p data.json -T application/json -c $(notdir $@) -n 2000 http://localhost:8000/api/v1/money/transfer

newman:
	newman run moneyservice.postman_collection.json

verify: build lint test check-arch

bench:
	k6 run ./bench/script.js
