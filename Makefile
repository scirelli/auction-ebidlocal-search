SHELL:=/usr/bin/env bash
.EXPORT_ALL_VARIABLES:

build: clean copy_configs copy_web copy_assets ./build/server
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
	@mkdir -p ./build/web/watchlists
	@cp -r ./web/user_templates/ ./build/template/
	@cp -r ./web/static/ ./build/web/static

copy_assets: assets
	@cp -r ./assets ./build/

.PHONY: requestNewUser
requestNewUser:
	@curl --request POST \
		--silent \
		--location \
		--header "Content-Type: application/json" \
	  	--data '{"name":"xyz"}' \
		localhost:8282/user | jq -r '.id'

.PHONY: requestNewWatchlist
requestNewWatchlist:
	curl --request POST \
		--include \
		--location \
		--header "Content-Type: application/json" \
		--data '{"name":"example", "list":["nintendo", "sega", "chainsaw", "turbografx", "playstation", "ps4", "ps3", "famicom", "macintosh"]}' \
		localhost:8282/user/$$EBID_USER/watchlist
	@echo ''
.PHONY: test
test:
	@go test ./...
