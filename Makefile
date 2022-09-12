NAME := prometheus-kafka-adapter
PACKAGE_NAME := github.com/Telefonica/prometheus-kafka-adapter
GO_VER := 1.19.1
LIBC_GO_VER := $(GO_VER)-buster
MUSL_GO_VER := $(GO_VER)-alpine

all: fmt test build

fmt:
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(MUSL_GO_VER) gofmt -l -w -s *.go

test:
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(MUSL_GO_VER) sh tools/testscript.sh vet
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(MUSL_GO_VER) sh tools/testscript.sh test

build: build-libc build-musl build-docker-image

build-libc:
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(LIBC_GO_VER) sh tools/buildscript.sh $(NAME)

build-musl:
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(MUSL_GO_VER) sh tools/buildscript.sh $(NAME)

build-docker-image:
	docker buildx build -t telefonica/prometheus-kafka-adapter:latest .

vendor-update:
	docker run --rm -e PACKAGE_NAME=$(PACKAGE_NAME) -v $(CURDIR):/app:z -w /app golang:$(MUSL_GO_VER) sh tools/vendorscript.sh
clean:
	rm -f $(NAME)-libc $(NAME)-musl
