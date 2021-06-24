NAME := prometheus-kafka-adapter
GO_VER := 1.16.5
LIBC_GO_VER := $(GO_VER)-buster
MUSL_GO_VER := $(GO_VER)-alpine

all: fmt test vet build

vendor:
	rm -rf vendor/
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(MUSL_GO_VER) apk add --no-cache gcc musl-dev && go mod vendor

fmt:
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(MUSL_GO_VER) gofmt -l -w -s *.go

test:
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(MUSL_GO_VER) sh tools/testscript.sh test

vet:
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(MUSL_GO_VER) sh tools/testscript.sh vet

vendor-update:
	rm -rf go.mod go.sum
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(MUSL_GO_VER) go mod init $(NAME)
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(MUSL_GO_VER) go mod tidy

build: build-libc build-musl

build-libc:
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(LIBC_GO_VER) sh tools/buildscript.sh $(NAME)

build-musl:
	docker run --rm -v $(CURDIR):/app:z -w /app golang:$(MUSL_GO_VER) sh tools/buildscript.sh $(NAME)
	# only build docker with the musl libraries 'cause we use an alpine container
	docker build .

clean:
	rm -f $(NAME)-libc $(NAME)-musl
