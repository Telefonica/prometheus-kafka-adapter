NAME := prometheus-kafka-adapter
GO_VER := 1.16.5
BUILD_GO_VER := $(GO_VER)-buster
LIBC_GO_VER := $(GO_VER)-buster
MUSL_GO_VER := $(GO_VER)-alpine

all: fmt vet test build

fmt:
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(BUILD_GO_VER) gofmt -l -w -s *.go

test:
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(BUILD_GO_VER) go test ./...

vet-libc:
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(LIBC_GO_VER) go vet -mod=vendor ./*.go

vendor-update: vendor-init tidy vendor

vendor-init:
	rm -rf go.mod go.sum vendor/
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(BUILD_GO_VER) go mod init $(NAME)

tidy:
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(BUILD_GO_VER) go mod tidy

vendor:
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(BUILD_GO_VER) go mod vendor


build: build-libc build-musl

build-libc:
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(LIBC_GO_VER) go build -mod=vendor -o $(NAME)-libc ./...
	docker build --build-arg binary=$(NAME)-libc .

build-musl:
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(MUSL_GO_VER) go build -tags musl -o $(NAME)-musl ./...
	docker build --build-arg binary=$(NAME)-musl .

clean:
	rm -f $(NAME)-libc $(NAME)-musl
