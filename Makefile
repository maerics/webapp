all: test

test: fmt vet
	go test ./...

build:
	go build .

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -f ./webapp ./*.db
