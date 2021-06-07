NAME := prometheus-kafka-adapter
GO_VER := 1.16.5
LIBC_GO_VER := $(GO_VER)-buster
MUSL_GO_VER := $(GO_VER)-alpine

all: build

fmt:
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(BUILD_GO_VER) gofmt -l -w -s *.go

vendor-update:
	rm -rf go.mod go.sum
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(BUILD_GO_VER) go mod init $(NAME)
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(BUILD_GO_VER) go mod tidy

build: build-libc build-musl

build-libc:
	#docker run --rm -v $(CURDIR):/app:z -w /app golang:$(LIBC_GO_VER) sh build.sh
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(LIBC_GO_VER) go test ./...
	rm -rf vendor/
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(LIBC_GO_VER) go mod vendor
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(LIBC_GO_VER) go vet -mod=vendor ./*.go
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(LIBC_GO_VER) go build -mod=vendor -ldflags='-w -s -extldflags "-static"' -o $(NAME) ./...
	rm -rf vendor/
	docker build --build-arg binary=$(NAME)-libc .

build-musl:
	#docker run --rm -v $(CURDIR):/app:z -w /app golang:$(MUSL_GO_VER) sh build.sh
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(MUSL_GO_VER) go test -tags musl ./...
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(MUSL_GO_VER) go vet -tags musl ./*.go
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(MUSL_GO_VER) go build -ldflags='-w -s -extldflags "-static"' -tags musl -o $(NAME) ./...
	docker build --build-arg binary=$(NAME)-musl .

clean:
	rm -f $(NAME)-libc $(NAME)-musl
