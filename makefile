SHELL:=/usr/bin/env bash
.EXPORT_ALL_VARIABLES:

build: clean copy_configs copy_web copy_assets ./build/server ## Build the project
	@echo 'Done'

ebidlocal: cmd/scanner/example.go
	@go build -o build/ebidlocal cmd/example.go

./build/server: cmd/server/main.go cmd/server/appConfig.go
	@go build -o build/server cmd/server/main.go cmd/server/appConfig.go

runEbid: ebidlocal
	@./build/ebidlocal | tee /tmp/index.html | md5 | tee /tmp/index.html.md5

run: ./build/server
	@cd ./build && \
	./server --config-path=$(shell pwd)/build/configs/config.json

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

/tmp/user.id: requestNewUser

requestNewUser: ## Create a new user, for testing
	@curl --request POST \
		--silent \
		--location \
		--header "Content-Type: application/json" \
	  	--data '{"name":"xyz"}' \
		localhost:8282/user | jq -r '.id' | tee /tmp/user.id

.PHONY: requestNewWatchlist
requestNewWatchlist: /tmp/user.id ## Create a new watch list. Used for testing.
	$$EBID_USER=$$(cat /tmp/user.id)
	curl --request POST \
		--include \
		--location \
		--header "Content-Type: application/json" \
		--data '{"name":"example", "list":["nintendo", "sega", "chainsaw", "turbografx", "playstation", "ps4", "ps3", "famicom", "macintosh"]}' \
		localhost:8282/user/$$EBID_USER/watchlist
	@echo ''

.PHONY: test
test: ## Run all tests
	@go test ./...

.PHONY: cleanTmpUserId
cleanTmpUserId: /tmp/user.id ## Remove the generated user.id
	@rm -rf /tmp/user.id

.PHONY: clean ## Remove generated build files
clean: cleanTmpUserId
	@rm -rf ./build

help: ## show help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m\033[0m\n"} /^[$$()% a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
