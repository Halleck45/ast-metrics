.PHONY: install build

PROTOC_VERSION=24.4
ARCHITECTURE=linux-x86_64

install-dev:install-protobuff 
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
	rm -rf src/NodeType || true
	GOPATH=$(HOME)/go PATH=$$PATH:$(HOME)/go/bin ./bin/protoc --go_out=src proto/NodeType.proto
	mv src/github.com/halleck45/ast-metrics/NodeType src
	rm -rf src/github.com
	@echo "\e[34m\033[1mDONE \033[0m\e[39m\n"
build-go: # for local development and tests
	@echo "\e[34m\033[1m-> Building go binaries\033[0m\e[39m\n"
	export CGO_LDFLAGS="-Lbuild/libgit/libgit2-1.5.0/build/ -Wl,-rpath -Wl,\$ORIGIN/build/libgit/libgit2-1.5.0/build/"
	export CGO_CFLAGS="-Ibuild/libgit/libgit2-1.5.0/build/"
	go build -o bin/ast-metrics
	@echo "\e[34m\033[1mDONE \033[0m\e[39m\n"
build-release:
	@echo "\e[34m\033[1m-> Building go binaries for supported platforms\033[0m\e[39m\n"
	rm -Rf dist || true
	go install github.com/goreleaser/goreleaser@latest
	GOPATH=$(HOME)/go PATH=$$PATH:$(HOME)/go/bin goreleaser build --snapshot
	@echo "\e[34m\033[1mDONE \033[0m\e[39m\n"



install-build: install-git
install-git: install-libgit bindata-bindata
install-libgit:
	@echo "\e[34m\033[1m-> Compiling libgit\033[0m\e[39m\n"
	rm -Rf build/libgit || true
	mkdir -p build/libgit
	cd build/libgit && curl -L https://github.com/libgit2/libgit2/archive/refs/tags/v1.5.0.tar.gz -o libgit2-1.5.0.tar.gz
	cd build/libgit && tar -xzf libgit2-1.5.0.tar.gz
	#cd build/libgit/libgit2-1.5.0 && mkdir build && cd build && cmake .. -DBUILD_CLAR=OFF -DTHREADSAFE=ON -DBUILD_SHARED_LIBS=OFF -DCMAKE_BUILD_TYPE=Release
	#cd build/libgit/libgit2-1.5.0/build/ && make 
	#cd build/libgit/libgit2-1.5.0/build/ && sudo make install 
	cd build/libgit/libgit2-1.5.0 && mkdir build && cd build && cmake -DTHREADSAFE=ON -DBUILD_CLAR=OFF -DCMAKE_BUILD_TYPE="RelWithDebInfo" .. && make && sudo make install
	cd build/libgit/libgit2-1.5.0/build/ && sudo ldconfig
bindata-bindata:
	@echo "\e[34m\033[1m-> Embedding libgit to current binary\033[0m\e[39m\n"
	rm -Rf build/libgit2/build/src build/libgit2/build/tests
	go install github.com/jteeuwen/go-bindata/...
	GOPATH=$(HOME)/go PATH=$$PATH:$(HOME)/go/bin go-bindata build/libgit/libgit2-1.5.0/build

	@echo "\e[34m\033[1mDONE \033[0m\e[39m\n"
	@echo "Remember to add build/libgit/libgit2-1.5.0/build to your LD_LIBRARY_PATH if you want to test"


build: install build-protobuff build-deps-libgit-embed build-release
	@echo "\n\e[42m  BUILD FINISHED  \e[49m\n"

test: test-go
test-go:
	@echo "\e[34m\033[1m-> Running tests\033[0m\e[39m\n"
	go test ./...
	@echo "\e[34m\033[1mDONE \033[0m\e[39m\n"





