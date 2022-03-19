all: lint test build

test:
	go mod tidy
	go test -timeout=10s -race -benchmem ./...

build:
	go build -o ftxgo cmd/main.go

lint: bin/golangci-lint
	go fmt ./...
	go vet ./...
	bin/golangci-lint -c .golangci.yml run ./...

bin/golangci-lint:
	wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.42.0

setup: bin/golangci-lint
	go mod download
