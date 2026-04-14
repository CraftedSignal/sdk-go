.PHONY: test lint coverage vet build clean

test:
	go test -race ./...

coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run --config .github/golangci.yml

vet:
	go vet ./...

build:
	CGO_ENABLED=0 go build ./...

clean:
	rm -f coverage.out coverage.html
