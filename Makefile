all: test

test: fmt vet
	go test ./...

build:
	go build \
		-ldflags " \
			-X 'webapp/cmd.BuildDirty=$(shell git status --short | sed "s/^ //" | tr "\n" ,)' \
			-X 'webapp/cmd.BuildBranch=$(shell git rev-parse --abbrev-ref HEAD)' \
			-X 'webapp/cmd.BuildVersion=$(shell git rev-parse HEAD)' \
			-X 'webapp/cmd.BuildTimestamp=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')' \
			" \
		.

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -f ./webapp ./*.db
