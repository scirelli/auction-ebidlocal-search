SHELL:=/usr/bin/env bash
.EXPORT_ALL_VARIABLES:

all: test

build: clean copy_configs copy_web copy_assets ./build/server ./build/scanner ## Build the project
	@echo 'Done'

./build/scanner: cmd/scanner/main.go cmd/scanner/appConfig.go
	@go build -o build/scanner cmd/scanner/main.go cmd/scanner/appConfig.go

./build/server: cmd/server/main.go cmd/server/appConfig.go
	@go build -o build/server cmd/server/main.go cmd/server/appConfig.go

scanner: ./build/scanner ## Run just the scanner
	@cd ./build && \
	./scanner --config-path=$(shell pwd)/build/configs/config.json

server: ./build/server ## Run just the webserver
	@cd ./build && \
	./server --config-path=$(shell pwd)/build/configs/config.json

./build/.running:
	make server &
	make scanner &
	touch ./build/.running

run: ./build/.running ## Run the server and scanner

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
		--data '{"name": "Steve C.", "email": "scirelli+ebidlocal@gmail.com"}' \
		localhost:8282/user | jq -r '.id' | tee /tmp/user.id

.PHONY: requestNewWatchlist
requestNewWatchlist: /tmp/user.id ## Create a new watch list. Used for testing.
	@EBID_USER=$$(cat /tmp/user.id) ; \
	curl --request POST \
		--include \
		--location \
		--header "Content-Type: application/json" \
		--data '{"name":"example", "list":["nintendo", "sega", "chainsaw", "turbografx", "playstation", "ps4", "ps3", "famicom", "macintosh", "xbox", "tv", "dreamcast", "psp", "vita", "lawnmower"]}' \
		localhost:8282/user/$$EBID_USER/watchlist
	@echo ''

.PHONY: requestNewWatchlist2
requestNewWatchlist2: /tmp/user.id ## Create a new watch list. Used for testing.
	@EBID_USER=$$(cat /tmp/user.id) ; \
	curl --request POST \
		--include \
		--location \
		--header "Content-Type: application/json" \
		--data '{"name":"example", "list":["nintendo", "sega", "chainsaw", "turbografx", "playstation", "ps4", "ps3", "famicom", "macintosh", "tv"]}' \
		localhost:8282/user/$$EBID_USER/watchlist
	@echo ''

.PHONY: test
test: ## Run all tests
	@go test ./...

.PHONY: vtest
vtest: ## Run all tests
	@go test -v -count=1 ./...

.PHONY: cleanTmpUserId
cleanTmpUserId: /tmp/user.id ## Remove the generated user.id
	@rm -rf /tmp/user.id

.PHONY: clean
clean: cleanTmpUserId  ## Remove generated build files
	@rm -rf ./build
	@go clean -testcache

# .PHONY: help
# help: ## Show help message
# 	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m\033[0m\n"} /^[$$()% a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: help
help: ## Show help message
	@grep -E '^[[:alnum:]_-]+[[:blank:]]?:.*##' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
