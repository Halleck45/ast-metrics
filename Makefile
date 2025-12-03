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
	rm -rf pb || true
	mkdir -p pb
	GOPATH=$(HOME)/go PATH=$$PATH:$(HOME)/go/bin ./bin/protoc --go_out=pb proto/NodeType.proto
	mv pb/github.com/halleck45/ast-metrics/pb/NodeType.pb.go pb/ || true
	echo 'THIS DIRECTORY IS BUILT BY MAKEFILE (make build-protobuff)' > pb/README.md
	rm -rf pb/github.com || true
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
	@echo "\e[34m\033[1m-> Preparing example projects for development\033[0m\e[39m\n"
	rm -rf ./samples
	@echo "  Creating directories..."
	mkdir -p ./samples/rust
	mkdir -p ./samples/golang
	mkdir -p ./samples/php
	mkdir -p ./samples/python
	@echo "  Cloning Rust projects..."
	git clone --depth 1 https://github.com/rust-lang/regex.git ./samples/rust/regex || true
	git clone --depth 1 https://github.com/BurntSushi/jiff.git ./samples/rust/jiff || true
	git clone --depth 1 https://github.com/DNSCrypt/encrypted-dns-server.git ./samples/rust/encrypted-dns-server || true
	git clone --depth 1 https://github.com/tokio-rs/tokio.git ./samples/rust/tokio || true
	git clone --depth 1 https://github.com/actix/actix-web.git ./samples/rust/actix-web || true
	git clone --depth 1 https://github.com/rust-lang/cargo.git ./samples/rust/cargo || true
	git clone --depth 1 https://github.com/rust-analyzer/rust-analyzer.git ./samples/rust/rust-analyzer || true
	git clone --depth 1 https://github.com/rust-lang/rust-clippy.git ./samples/rust/rust-clippy || true
	git clone --depth 1 https://github.com/serde-rs/serde.git ./samples/rust/serde || true
	git clone --depth 1 https://github.com/hyperium/hyper.git ./samples/rust/hyper || true
	git clone --depth 1 https://github.com/rust-lang/rust.git ./samples/rust/rust || true
	git clone --depth 1 https://github.com/rust-lang/rustfmt.git ./samples/rust/rustfmt || true
	git clone --depth 1 https://github.com/rayon-rs/rayon.git ./samples/rust/rayon || true
	git clone --depth 1 https://github.com/clap-rs/clap.git ./samples/rust/clap || true
	@echo "  Cloning Go projects..."
	git clone --depth 1 https://github.com/DNSCrypt/dnscrypt-proxy.git ./samples/golang/dnscrypt-proxy || true
	git clone --depth 1 https://github.com/gin-gonic/gin.git ./samples/golang/gin || true
	git clone --depth 1 https://github.com/kubernetes/kubernetes.git ./samples/golang/kubernetes || true
	git clone --depth 1 https://github.com/docker/docker.git ./samples/golang/docker || true
	git clone --depth 1 https://github.com/etcd-io/etcd.git ./samples/golang/etcd || true
	git clone --depth 1 https://github.com/prometheus/prometheus.git ./samples/golang/prometheus || true
	git clone --depth 1 https://github.com/grafana/grafana.git ./samples/golang/grafana || true
	git clone --depth 1 https://github.com/hashicorp/terraform.git ./samples/golang/terraform || true
	git clone --depth 1 https://github.com/hashicorp/consul.git ./samples/golang/consul || true
	git clone --depth 1 https://github.com/grpc/grpc-go.git ./samples/golang/grpc-go || true
	git clone --depth 1 https://github.com/golang/go.git ./samples/golang/go || true
	git clone --depth 1 https://github.com/cockroachdb/cockroach.git ./samples/golang/cockroach || true
	git clone --depth 1 https://github.com/gorilla/mux.git ./samples/golang/gorilla-mux || true
	git clone --depth 1 https://github.com/spf13/cobra.git ./samples/golang/cobra || true
	git clone --depth 1 https://github.com/spf13/viper.git ./samples/golang/viper || true
	git clone --depth 1 https://github.com/uber-go/zap.git ./samples/golang/zap || true
	git clone --depth 1 https://github.com/stretchr/testify.git ./samples/golang/testify || true
	git clone --depth 1 https://github.com/go-redis/redis.git ./samples/golang/go-redis || true
	@echo "  Cloning PHP projects..."
	git clone --depth 1 https://github.com/WordPress/wordpress-develop.git ./samples/php/wordpress-develop || true
	git clone --depth 1 https://github.com/symfony/messenger.git ./samples/php/messenger || true
	git clone --depth 1 https://github.com/symfony/symfony.git ./samples/php/symfony || true
	git clone --depth 1 https://github.com/laravel/framework.git ./samples/php/laravel-framework || true
	git clone --depth 1 https://github.com/doctrine/orm.git ./samples/php/doctrine-orm || true
	git clone --depth 1 https://github.com/phpunit/phpunit.git ./samples/php/phpunit || true
	git clone --depth 1 https://github.com/monolog/monolog.git ./samples/php/monolog || true
	git clone --depth 1 https://github.com/guzzle/guzzle.git ./samples/php/guzzle || true
	git clone --depth 1 https://github.com/twigphp/Twig.git ./samples/php/twig || true
	git clone --depth 1 https://github.com/composer/composer.git ./samples/php/composer || true
	git clone --depth 1 https://github.com/phpstan/phpstan.git ./samples/php/phpstan || true
	git clone --depth 1 https://github.com/php-fig/fig-standards.git ./samples/php/fig-standards || true
	git clone --depth 1 https://github.com/phpredis/phpredis.git ./samples/php/phpredis || true
	git clone --depth 1 https://github.com/reactphp/react.git ./samples/php/reactphp || true
	git clone --depth 1 https://github.com/yiisoft/yii2.git ./samples/php/yii2 || true
	git clone --depth 1 https://github.com/zendframework/zendframework.git ./samples/php/zendframework || true
	git clone --depth 1 https://github.com/nette/nette.git ./samples/php/nette || true
	@echo "  Cloning Python projects..."
	git clone --depth 1 https://github.com/WordPress/openverse.git ./samples/python/openverse || true
	@echo "\e[34m\033[1mDONE \033[0m\e[39m\n"
	@echo "\e[32m\033[1mExample projects prepared in ./samples/\033[0m\e[39m\n"

clean:
	rm -rf bin dist build protoc.zip coverage.txt || true
