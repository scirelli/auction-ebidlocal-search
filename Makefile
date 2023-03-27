SHELL:=/usr/bin/env bash
.EXPORT_ALL_VARIABLES:

# EMAIL_PASSWORD=   # required in environment to send passwords

ifeq (, $(shell which jq))
	$(error "No jq in $(PATH), consider doing apt-get install jq")
endif

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
	@cp -r ./web/static/ ./build/web/static

copy_assets: assets
	@cp -r ./assets ./build/

/tmp/user.id:
	@$(MAKE) requestNewUser

requestNewUser: ## Create a new user, for testing
	@curl --request POST \
		--silent \
		--location \
		--header "Content-Type: application/json" \
		--data '{"name": "Steve C.", "email": "scirelli+ebidlocal@gmail.com"}' \
		localhost:8282/user | jq -r '.id' | tee /tmp/user.id

.PHONY: requestNewWatchlist
requestNewWatchlist: /tmp/user.id ## Create a new watch list. Used for testing.
	@EBID_USER="$$(cat /tmp/user.id)" ; \
	curl --request POST \
		--include \
		--location \
		--header "Content-Type: application/json" \
		--data '{"name":"example-retro", "list":["nintendo", "sega", "turbografx", "playstation", "ps4", "ps3", "famicom", "macintosh", "xbox", "tv", "dreamcast", "psp", "vita", "commodore", "turboexpress", "turbo", "amiga", "tandy"]}' \
		localhost:8282/user/$$EBID_USER/watchlist
	@echo ''

.PHONY: requestDeleteWatchlist
requestDeleteWatchlist: /tmp/user.id ## Create a new watch list. Used for testing.
	@EBID_USER="$$(cat /tmp/user.id)" ; \
	curl --request DELETE \
		--include \
		--location \
		--header "Content-Type: application/json" \
		--data '{"name":"example"}' \
		localhost:8282/user/$$EBID_USER/watchlist
	@echo ''

.PHONY: requestNewWatchlist2
requestNewWatchlist2: /tmp/user.id ## Create a new watch list. Used for testing.
	@EBID_USER="$$(cat /tmp/user.id)" ; \
	curl --request POST \
		--include \
		--location \
		--header "Content-Type: application/json" \
		--data '{"name":"example2-household", "list":["recliner", "pool", "chainsaw", "saw", "mower", "lawnmower", "scuba", "tarp", "bed", "stair", "stepper", "climber", "headphones", "drill", "drillpress", "press", "tent"]}' \
		localhost:8282/user/$$EBID_USER/watchlist
	@echo ''

.PHONY: requestNewWatchlist3
requestNewWatchlist3: /tmp/user.id ## Create a new watch list. Used for testing.
	@EBID_USER="$$(cat /tmp/user.id)" ; \
	curl --request POST \
		--include \
		--location \
		--header "Content-Type: application/json" \
		--data '{"name":"example-board-games", "list":["chess", "monopoly", "candyland"]}' \
		localhost:8282/user/$$EBID_USER/watchlist
	@echo ''

.PHONY: requestNewWatchlist4
requestNewWatchlist4: /tmp/user.id ## Create a new watch list. Used for testing.
	@EBID_USER="$$(cat /tmp/user.id)" ; \
	curl --request POST \
		--include \
		--location \
		--header "Content-Type: application/json" \
		--data '{"name":"example-holiday", "list":["christmas", "xmas", "x-mas", "halloween", "fall", "decorations"]}' \
		localhost:8282/user/$$EBID_USER/watchlist
	@echo ''

.PHONY: requestNewWatchlist5
requestNewWatchlist5: /tmp/user.id ## Create a new watch list. Used for testing.
	@EBID_USER="$$(cat /tmp/user.id)" ; \
	curl --request POST \
		--include \
		--location \
		--header "Content-Type: application/json" \
		--data '{"name":"example-music", "list":["sheet", "violin", "cello", "music", "piano"]}' \
		localhost:8282/user/$$EBID_USER/watchlist
	@echo ''

.PHONY: test
test: ## Run all tests
	@go test ./...

.PHONY: vtest
vtest: ## Run all tests with verbose flag set
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

.PHONY: scannerServiceLog ## Follow the scanner service logs
scannerServiceLog:
	sudo journalctl -f -u watchlist-scanner.service

.PHONY: serverServiceLog
serverServiceLog: ## Follow the server service logs
	sudo journalctl -f -u watchlist-http.service

.PHONY: verifyUser
verifyUser: /tmp/user.id /tmp/user.nonce
	EBID_USER="$$(cat /tmp/user.id)" ; \
	NONCE="$$(cat /tmp/user.nonce)" ; \
	curl --request GET \
		--header "Content-type: application/json" \
		http://localhost:8282/user/$$EBID_USER/verify/$$NONCE

/tmp/user.nonce: /tmp/user.id
	@EBID_USER="$$(cat /tmp/user.id)" ; \
	cat /tmp/web/user/$$EBID_USER/data.json | jq -r .verifyToken > /tmp/user.nonce \

.PHONY: sendVerification
sendVerification: /tmp/user.id ## Send verification email.
	EBID_USER="$$(cat /tmp/user.id)" ; \
	curl --request PUT \
		--header "Content-type: application/json" \
		http://localhost:8282/user/$$EBID_USER/verify/send

.PHONY: help
help: ## Show help message
	@grep -E '^[[:alnum:]_-]+[[:blank:]]?:.*##' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
