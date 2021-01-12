SHELL:=/usr/bin/env bash

build: clean copy_configs copy_web ./build/server
	@echo 'Done'

ebidlocal: cmd/example.go
	@go build -o build/ebidlocal cmd/example.go

./build/server: cmd/main.go
	@go build -o build/server cmd/main.go cmd/appConfig.go

runEbid: ebidlocal
	@./build/ebidlocal | tee /tmp/index.html | md5 | tee /tmp/index.html.md5 

run: ./build/server
	@cd ./build && \
	./server --config-path=$(shell pwd)/build/configs/config.json

.PHONY: clean
clean:
	@rm -rf ./build

copy_configs: configs
	@mkdir -p ./build/configs
	@cp -r ./configs/*.json ./build/configs/ || :

copy_web: web
	@mkdir -p ./build/web/user
	@cp -r ./web/user_templates/ ./build/template/
	@cp -r ./web/static/ ./build/web/static

.PHONY: requestNewUser
requestNewUser:
	curl --request POST \
		--include \
		--header "Content-Type: application/json" \
	  	--data '{"username":"xyz"}' \
		localhost:8282/user
	@echo ''

.PHONY: test
test:
	@go test ./...
