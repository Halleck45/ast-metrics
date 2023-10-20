.PHONY: install build

install:
	cd runner/php && composer install
build: install
	go build