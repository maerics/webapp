all: test

test:
	go test ./...

build:
	go build .

tidy:
	go mod tidy
