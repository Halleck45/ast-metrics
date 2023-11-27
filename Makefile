.PHONY: install build

PROTOC_VERSION=24.4
ARCHITECTURE=linux-x86_64

install:install-protobuff
	@echo "\e[34m\033[1mDONE \033[0m\e[39m\n"
install-protobuff:
	@echo "\e[34m\033[1m-> Downloading protobuff\033[0m\e[39m\n"
	mkdir -p bin
	rm -Rf bin/protoc include readme.txt || true
	rm protoc.zip || true
	curl --silent -L "https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-$(ARCHITECTURE).zip" -o "protoc.zip"
	unzip protoc.zip
	rm -Rf protoc.zip include readme.txt || true
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/golang/protobuf/protoc-gen-go
	@echo "\e[34m\033[1mDONE \033[0m\e[39m\n"
build-protobuff:
	@echo "\e[34m\033[1m-> Building protobuff\033[0m\e[39m\n"
	mv src/github.com/halleck45/ast-metrics/NodeType src
	rm -rf src/github.com
	@echo "\e[34m\033[1mDONE \033[0m\e[39m\n"
build-go: # for local development and tests
	@echo "\e[34m\033[1m-> Building go binaries\033[0m\e[39m\n"
	go build -o bin/ast-metrics
	@echo "\e[34m\033[1mDONE \033[0m\e[39m\n"
build-release:
	@echo "\e[34m\033[1m-> Building go binaries for supported platforms\033[0m\e[39m\n"
	rm -Rf dist || true
	go install github.com/goreleaser/goreleaser@latest
	GOPATH=$(HOME)/go PATH=$$PATH:$(HOME)/go/bin goreleaser build --snapshot
	@echo "\e[34m\033[1mDONE \033[0m\e[39m\n"
build: install build-protobuff build-release
	@echo "\n\e[42m  BUILD FINISHED  \e[49m\n"

test: test-go
test-go:
	@echo "\e[34m\033[1m-> Running tests\033[0m\e[39m\n"
	go test ./...
	@echo "\e[34m\033[1mDONE \033[0m\e[39m\n"