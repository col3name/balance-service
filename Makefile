.PHONY: clean test lint

all: clean bin/money test lint check-arch

clean:
	rm -rf bin/*

bin/%:
	GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -o ./bin/$(notdir $@) ./cmd/$(notdir $@)

test:
	go test ./...

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
	newman run postman.json