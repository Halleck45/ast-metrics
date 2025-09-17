.PHONY: install build monkey-test

PROTOC_VERSION=24.4
ARCHITECTURE=linux-x86_64

bin/protoc:
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
build-protobuff: bin/protoc
	@echo "\e[34m\033[1m-> Building protobuff\033[0m\e[39m\n"
	rm -rf internal/nodetype || true
	GOPATH=$(HOME)/go PATH=$$PATH:$(HOME)/go/bin ./bin/protoc --go_out=internal proto/NodeType.proto
	mv internal/github.com/halleck45/ast-metrics/NodeType internal/nodetype || true
	rm -rf internal/github.com || true
	@echo "\e[34m\033[1mDONE \033[0m\e[39m\n"
build-go: # for local development and tests
	@echo "\e[34m\033[1m-> Building go binaries\033[0m\e[39m\n"
	go build -o bin/ast-metrics ./cmd/ast-metrics
	@echo "\e[34m\033[1mDONE \033[0m\e[39m\n"
build-release:
	@echo "\e[34m\033[1m-> Building go binaries for supported platforms\033[0m\e[39m\n"
	rm -Rf dist || true
	go install github.com/goreleaser/goreleaser@latest
	GOPATH=$(HOME)/go PATH=$$PATH:$(HOME)/go/bin goreleaser build --snapshot
	@echo "\e[34m\033[1mDONE \033[0m\e[39m\n"
build: build-protobuff build-go
	@echo "\n\e[42m  BUILD FINISHED  \e[49m\n"

test:
	@echo "\e[34m\033[1m-> Running tests\033[0m\e[39m\n"
	go clean -testcache
	find . -type d  -iname ".ast-metrics-cache" -exec rm -rf "{}" \; || true
	go test ./...
	@echo "\e[34m\033[1mDONE \033[0m\e[39m\n"

# monkey test: download random PHP and Go packages from top 100 and analyze them
monkey-test:
	@echo "\e[34m\033[1m-> Monkey testing\033[0m\e[39m\n"
	bash scripts/monkey-test.sh
	@echo "\e[34m\033[1mDONE \033[0m\e[39m\n"

# profiling
profile:
	go run ./cmd/ast-metrics a --non-interactive --profile .
	go tool pprof -png  ast-metrics.cpu
	go tool pprof -png  ast-metrics.mem


dev-prepare-examples:
	rm -rf /tmp/samples
	mkdir -p /tmp/samples/rust
	git clone git@github.com:rust-lang/regex.git /tmp/samples/rust/regex
	git clone git@github.com:BurntSushi/jiff.git /tmp/samples/rust/jiff
	git clone git@github.com:DNSCrypt/encrypted-dns-server.git /tmp/samples/rust/encrypted-dns-server
	mkdir -p /tmp/samples/golang
	git clone git@github.com:DNSCrypt/dnscrypt-proxy.git /tmp/samples/golang/dnscrypt-proxy
	mkdir -p /tmp/samples/php
	git clone git@github.com:WordPress/wordpress-develop.git /tmp/samples/php/wordpress-develop
	git clone git@github.com:symfony/messenger.git /tmp/samples/php/messenger
	mkdir -p /tmp/samples/python
	git clone git@github.com:WordPress/openverse.git /tmp/samples/python/openverse

clean:
	rm -rf bin dist build protoc.zip coverage.txt || true
