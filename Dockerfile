FROM golang:alpine AS builder
RUN apk update
RUN apk add --no-cache gcc git make musl-dev

WORKDIR /src
ADD go.mod go.sum /src/
RUN go mod download

COPY . .
RUN make build

FROM alpine
WORKDIR /
COPY --from=builder /src/webapp /webapp
USER 1001:1001
CMD ["/webapp", "version"]
