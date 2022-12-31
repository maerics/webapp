all: test

test:
	go test ./...

build:
	go build \
		-ldflags " \
			-X 'webapp/cmd.BuildBranch=$(shell git rev-parse --abbrev-ref HEAD)'  \
			-X 'webapp/cmd.BuildVersion=$(shell git rev-parse HEAD)'              \
			-X 'webapp/cmd.BuildTimestamp=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')' \
			" \
		.

clean:
	rm -f ./webapp ./*.db
