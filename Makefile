DOCKER_REGISTRY_PREFIX=gcr.io/pioneering-fuze-373200

all: test

test: fmt vet
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -f ./webapp ./*.db

build:
	$(eval BRANCH_NAME := $(shell git rev-parse --abbrev-ref HEAD))
	go build -ldflags "-X 'webapp/web.Branch=${BRANCH_NAME}'" .

release: # ensure-no-local-changes
	$(eval VERSION := $(shell git rev-parse HEAD))
	$(eval IMAGE_NAME := "${DOCKER_REGISTRY_PREFIX}/webapp:${VERSION}")
	docker build -f Dockerfile -t ${IMAGE_NAME} .
	docker run --rm -it ${IMAGE_NAME} /webapp --version
	docker push ${IMAGE_NAME}
	docker image rm ${IMAGE_NAME}

ensure-no-local-changes:
	@if [ "$(shell git status -s)" != "" ]; then \
		git status -s; \
		echo "FATAL: refusing to release with local changes; see git status."; \
		exit 1; \
	fi
