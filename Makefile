NAME := prometheus-kafka-adapter
PACKAGE_NAME := github.com/Telefonica/prometheus-kafka-adapter
GO_VER := 1.17.3
LIBC_GO_VER := $(GO_VER)-buster
MUSL_GO_VER := $(GO_VER)-alpine

all: fmt test build

fmt:
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(MUSL_GO_VER) gofmt -l -w -s *.go

test:
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(MUSL_GO_VER) sh tools/testscript.sh vet
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(MUSL_GO_VER) sh tools/testscript.sh test

build: build-libc build-musl

build-libc:
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(LIBC_GO_VER) sh tools/buildscript.sh $(NAME)

build-musl:
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(MUSL_GO_VER) sh tools/buildscript.sh $(NAME)

vendor-update:
	rm -rf go.mod go.sum vendor/
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(MUSL_GO_VER) go mod init $(PACKAGE_NAME)
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(MUSL_GO_VER) go mod tidy
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(MUSL_GO_VER) apk add --no-cache gcc musl-dev && go mod vendor

clean:
	rm -f $(NAME)-libc $(NAME)-musl
